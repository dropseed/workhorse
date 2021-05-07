import re
import subprocess
import datetime


def is_dirty():
    output = subprocess.check_output(["git", "status", "--porcelain"])
    return bool(output.strip())


def create_branch(name):
    subprocess.check_call(["git", "checkout", "-b", name])


def delete_branch(name):
    subprocess.check_call(["git", "branch", "-D", name])


def repo_from_remote():
    output = (
        subprocess.check_output(["git", "remote", "get-url", "origin"])
        .decode("utf-8")
        .strip()
    )
    return re.search(r"/([^/]+/[^/]+?)(\.git)?$", output)[1]


def add_commit(path, message):
    subprocess.check_call(["git", "add", path])
    subprocess.check_call(["git", "commit", "-m", message])


def push(branch):
    subprocess.check_call(
        ["git", "push", "--force", "--set-upstream", "origin", branch]
    )


def checkout(ref):
    subprocess.check_call(["git", "checkout", ref])


def last_commit_message():
    return (
        subprocess.check_output(["git", "show", "-s", "--format=%s"])
        .decode("utf-8")
        .strip()
    )


def last_commit_files_added():
    return (
        subprocess.check_output(
            ["git", "diff", "HEAD^", "HEAD", "--name-only", "--diff-filter", "A"]
        )
        .decode("utf-8")
        .strip()
        .splitlines()
    )


def commit_datetime_of_path(path):
    try:
        date_string = (
            subprocess.check_output(
                ["git", "log", "-1", "--format=%cd", "--date=iso-strict", path]
            )
            .decode("utf-8")
            .strip()
        )
        return datetime.datetime.fromisoformat(date_string)
    except (ValueError, subprocess.subprocess.CalledProcessError):
        return None
