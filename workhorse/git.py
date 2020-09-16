import subprocess


def is_dirty():
    output = subprocess.check_output(["git", "status", "--porcelain"])
    return bool(output.strip())
