import re
import subprocess


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
