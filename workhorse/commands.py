import subprocess


def shell(run, stdin=None):
    output = subprocess.run(run, shell=True, stdin=stdin)


def sleep(duration):
    from time import sleep
    sleep(duration)


available_pull_commands = {
    "sleep": sleep,
    "shell": shell,
}

available_repo_commands = {
    "sleep": sleep,
    "shell": shell,
}
