import os
import re
import base64
import tempfile
import subprocess

from cached_property import cached_property

from .api import session
from .user_input.expressions import Expression
from .user_input import template
from .exceptions import RetryException


def get_api_url(url):
    """Get a API URL (without base) for a given HTML URL or API URL"""
    repo_match = re.search(r"/([^/]+)/([^/]+)$", url)
    issue_match = re.search(r"/([^/]+)/([^/]+)/(pull|issue)/(\d+)$", url)
    if issue_match:
        return f"repos/{issue_match[1]}/{issue_match[2]}/{issue_match[3]}s/{issue_match[4]}"
    elif repo_match:
        return f"repos/{repo_match[1]}/{repo_match[2]}"


class Target:
    """
    This object itself will get injected into the context,
    so anything with _ is inaccessible there
    """
    def __init__(self, type, url):
        self._type = type.rstrip("s")  # allow with or without s... repos, repo, pulls, etc.
        self._url = url
        self._api_url = get_api_url(url)
        self._cwd = os.getcwd()

        self.repo = None
        self.pull = None

    def _load(self, repo=None, pull=None):
        if self._type == "repo":
            self.repo = self
            self.pull = pull
            if self.pull:
                self.pull._load(repo=self.repo)

        elif self._type == "pull":
            self.pull = self
            if repo:
                self.repo = repo
            else:
                self.repo = Target("repo", self._data["head"]["repo"]["html_url"])
                self.repo._load()  # don't pass pull because don't need recursive

        # make the API props directly accessible with .
        # (doesn't do nested context stuff like pullapprove does)
        self.__dict__.update(self._data)

    @cached_property
    def _data(self):
        response = session.get(self._api_url)
        response.raise_for_status()
        return response.json()

    def _clear_cache(self):
        if "_data" in self.__dict__:
            del self.__dict__["_data"]
        if self.pull and "_data" in self.pull.__dict__:
            del self.pull.__dict__["_data"]
        if self.repo and "_data" in self.repo.__dict__:
            del self.repo.__dict__["_data"]

    def _render_markdown(self, s):
        return template.render(s, self._get_context())

    def _get_context(self):
        return {
            "repo": self.repo,
            "pull": self.pull,
        }

    def _get_commands(self):
        ctx = {}
        prefix = "_cmd_"
        for k in dir(self):
            if k.startswith(prefix):
                ctx[k[len(prefix) :]] = getattr(self, k)
        return ctx

    def _run_command(self, name, params):
        cmd = getattr(self, f"_cmd_{name}")
        return cmd(**params)

    def _expression_result(self, s):
        return Expression(s, self._get_context()).compile()

    def file_contents(self, path):
        response = session.get(f"{self.repo._api_url}/contents/{path}")
        if response.status_code == 404:
            return ""
        response.raise_for_status()
        contents = base64.b64decode(response.json()["content"].encode("utf-8")).decode(
            "utf-8"
        )
        return contents

    def path_exists(self, path):
        # TODO could have PR behavior too?
        response = session.get(f"{self.repo._api_url}/contents/{path}")
        if response.status_code == 404:
            return False
        response.raise_for_status()
        return True

    def branch_exists(self, name):
        response = session.get(f"{self.repo._api_url}/branches/{name}")
        if response.status_code == 404:
            return False
        response.raise_for_status()
        return True

    # TODO could have more interactive shell env if you used decorators or something rather than prefix
    # - could also have ipdb command to drop in maybe? what would make it easier to develop? "run" cmd to skip planning...
    # and/or ability to confirm between steps, or write steps as you go? prompt for cmd, then args, then saves it if successful to a plan file
    def _cmd_shell(self, run, input=None, env={}):
        if isinstance(input, str):
            input = input.encode("utf-8")
        env["REPO"] = re.search("/?repos/([^/]+/[^/]+)", self.repo._api_url).groups()[0]
        merged_env = os.environ.copy()
        merged_env.update(env)
        subprocess.run(run, shell=True, env=merged_env, check=True, input=input, cwd=self._cwd)

    def _cmd_sleep(self, duration):
        from time import sleep

        sleep(duration)

    def _cmd_api(
        self, method, url=None, repo_url=None, json=None, params=None, headers=None
    ):
        if repo_url:
            base = re.search("(/?repos/[^/]+/[^/]+)", self.repo._api_url).groups()[0]
            request_url = base + "/" + repo_url.lstrip("/")
        else:
            request_url = url or self._api_url
        response = session.request(
            method.lower(), request_url, json=json, params=params, headers=headers
        )
        response.raise_for_status()

    def _cmd_clone(self, path=None, depth=1):
        if not path:
            path = os.path.join(
                tempfile.TemporaryDirectory(prefix="workhorse-").name, "repo"
            )
        args = []
        if depth:
            args += ["--depth", str(depth)]
        subprocess.check_call(["git", "clone"] + args + [self.repo._url, path])
        self._cwd = path  # for future subprocess calls

    def _cmd_wait(self, condition):
        self._clear_cache()
        self._load(repo=self.repo, pull=self.pull)
        result = self._expression_result(condition)
        if not result:
            raise RetryException(f"Wait condition not satisfied: {result}")

    def _cmd_create_pull(self, title, head, base, body="", draft=False):
        response = session.post(
            f"{self.repo._api_url}/pulls", json={"title": title, "head": head, "base": base, "body": body, "draft": draft}
        )
        response.raise_for_status()
        pull_url = response.json()["html_url"]

        # set self.pull for future context
        self._load(pull=Target("pull", pull_url))

    def _cmd_merge(self, merge_method=None):
        # TODO repo option?
        response = session.put(
            f"{self.pull._api_url}/merge", json={"merge_method": merge_method}
        )
        response.raise_for_status()

    def _cmd_close(self):
        # TODO repo option? number? what if issues are added?
        response = session.patch(self.pull._api_url, json={"state": "closed"})
        response.raise_for_status()

    def _cmd_delete_branch(self, name=None):
        if name:
            ref = name
        else:
            pull = self.pull._data
            ref = pull["head"]["ref"]

        response = session.delete(
            self.repo._data["git_refs_url"].replace("{/sha}", f"/heads/{ref}")
        )
        response.raise_for_status()

    def _cmd_create_branch(self, name):
        response = session.get(f"{self.repo._api_url}/commits", params={"ref": name})
        response.raise_for_status()
        latest_commit_sha = response.json()[0]["sha"]

        ref = f"refs/heads/{name}"
        response = session.post(
            f"{self.repo._api_url}/git/refs",
            json={
                "ref": ref,
                "sha": latest_commit_sha,
            }
        )
        response.raise_for_status()

    def _cmd_replace_in_file(self, file, find, replace, branch, message=None):
        file_url = f"{self.repo._api_url}/contents/{file}"

        response = session.get(file_url, params={"ref": branch})
        response.raise_for_status()
        original_contents = base64.b64decode(response.json()["content"].encode("utf-8")).decode(
            "utf-8"
        )
        original_sha = response.json()["sha"]

        updated_contents = original_contents.replace(find, replace)

        if original_contents.strip() == updated_contents.strip():
            raise Exception("No change in find replace")

        response = session.put(file_url, json={
            "message": message or f"Replace {find} with {replace}",
            "content": base64.b64encode(updated_contents.encode("utf-8")).decode("utf-8"),
            "sha": original_sha,
            "branch": branch,
        })
        response.raise_for_status()
