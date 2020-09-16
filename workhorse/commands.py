import subprocess

from .api import session
from .targets import get_api_url


def shell(target_url, run, stdin=None):
    output = subprocess.run(run, shell=True, stdin=stdin)


def sleep(target_url, duration):
    from time import sleep

    sleep(duration)


def merge_pull(target_url, merge_method=None):
    response = session.put(
        get_api_url(target_url), params={"merge_method": merge_method}
    )
    response.raise_for_status()


def update_pull(
    target_url, title=None, body=None, state=None, base=None, maintainer_can_modify=None
):
    data = {}

    if title is not None:
        data["title"] = title
    if body is not None:
        data["body"] = body
    if state is not None:
        data["state"] = state
    if base is not None:
        data["base"] = base
    if maintainer_can_modify is not None:
        data["maintainer_can_modify"] = maintainer_can_modify

    response = session.patch(get_api_url(target_url), json=data)
    response.raise_for_status()


def delete_pull_branch(target_url):
    pull_url = get_api_url(target_url)

    response = session.get(pull_url)
    response.raise_for_status()
    pull = response.json()
    ref = pull["head"]["ref"]

    response = session.delete(
        pull["head"]["repo"]["git_refs_url"].replace("{/sha}", f"heads/{ref}")
    )
    response.raise_for_status()


available_pull_commands = {
    "sleep": sleep,
    "shell": shell,
    "merge_pull": merge_pull,
    "update_pull": update_pull,
    "delete_pull_branch": delete_pull_branch,
}

available_repo_commands = {
    "sleep": sleep,
    "shell": shell,
}
