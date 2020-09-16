import subprocess


def is_dirty():
    output = subprocess.check_output(["git", "status", "--porcelain"])
    return bool(output.strip())


def create_branch(name):
    subprocess.check_call(["git", "checkout", "-b", name])


def delete_branch(name):
    subprocess.check_call(["git", "branch", "-D", name])
