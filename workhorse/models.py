import os
import re
import base64
import tempfile
import subprocess

from cached_property import cached_property

from .api import session
from .user_input.expressions import Expression


def get_api_url(url):
    """Get a API URL (without base) for a given HTML URL or API URL"""
    repo_match = re.search(r"/([^/]+)/([^/]+)$", url)
    issue_match = re.search(r"/([^/]+)/([^/]+)/(pull|issue)/(\d+)$", url)
    if issue_match:
        return f"repos/{issue_match[1]}/{issue_match[2]}/{issue_match[3]}s/{issue_match[4]}"
    elif repo_match:
        return f"repos/{repo_match[1]}/{repo_match[2]}"


def model_for_url(url):
    if re.search(r"/([^/]+)/([^/]+)/(pull|issue)/(\d+)$", url):
        return Pull(url)
        # issue eventually
    else:
        return Repo(url)


class BaseContextModel:
    def __init__(self, url):
        self.url = url
        self.api_url = get_api_url(url)
        self.cwd = os.getcwd()

    @cached_property
    def data(self):
        response = session.get(self.api_url)
        response.raise_for_status()
        return response.json()

    def get_filter_context(self):
        ctx = self.data
        prefix = "filter_"
        for k in dir(self):
            if k.startswith(prefix):
                ctx[k[len(prefix) :]] = getattr(self, k)
        return ctx

    def get_commands(self):
        ctx = {}
        prefix = "cmd_"
        for k in dir(self):
            if k.startswith(prefix):
                ctx[k[len(prefix) :]] = getattr(self, k)
        return ctx

    def matches_filter(self, filter):
        return Expression(filter, self.get_filter_context()).compile()

    def cmd_shell(self, run, input=None, env={}):
        if isinstance(input, str):
            input = input.encode("utf-8")
        env["REPO"] = re.search("/?repos/([^/]+/[^/]+)", self.api_url).groups()[0]
        subprocess.run(run, shell=True, env=env, check=True, input=input, cwd=self.cwd)

    def cmd_sleep(self, duration):
        from time import sleep

        sleep(duration)

    def cmd_api(
        self, method, url=None, repo_url=None, json=None, params=None, headers=None
    ):
        if repo_url:
            base = re.search("(/?repos/[^/]+/[^/]+)", self.api_url).groups()[0]
            request_url = base + "/" + repo_url.lstrip("/")
        else:
            request_url = url or self.api_url
        response = session.request(
            method.lower(), request_url, json=json, params=params, headers=headers
        )
        response.raise_for_status()


class Repo(BaseContextModel):
    def filter_file_contains(self, path, s):
        response = session.get(f"{self.api_url}/contents/{path}")
        if response.status_code == 404:
            return False
        response.raise_for_status()
        contents = base64.b64decode(response.json()["content"].encode("utf-8")).decode(
            "utf-8"
        )
        return s in contents

    def filter_path_exists(self, path):
        response = session.get(f"{self.api_url}/contents/{path}")
        if response.status_code == 404:
            return False
        response.raise_for_status()
        return True

    def filter_branch_exists(self, name):
        response = session.get(f"{self.api_url}/branches/{name}")
        if response.status_code == 404:
            return False
        response.raise_for_status()
        return True

    # TODO even this doesn't necessarily have to exist?
    # could clone in shell, if you have env vars...
    # need to set token if CI...
    def cmd_clone(self, path=None, depth=1):
        if not path:
            path = os.path.join(
                tempfile.TemporaryDirectory(prefix="workhorse-").name, "repo"
            )
        args = []
        if depth:
            args += ["--depth", str(depth)]
        subprocess.check_call(["git", "clone"] + args + [self.url, path])
        self.cwd = path  # for future subprocess calls


class Pull(BaseContextModel):
    def cmd_merge(self, merge_method=None):
        response = session.put(
            f"{self.api_url}/merge", json={"merge_method": merge_method}
        )
        response.raise_for_status()

    def cmd_close(self):
        response = session.patch(self.api_url, json={"state": "closed"})
        response.raise_for_status()

    def cmd_delete_branch(self):
        pull = self.data
        ref = pull["head"]["ref"]

        response = session.delete(
            pull["head"]["repo"]["git_refs_url"].replace("{/sha}", f"/heads/{ref}")
        )
        response.raise_for_status()
