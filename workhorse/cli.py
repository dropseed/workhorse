import os
import json
import re

import yaml
import click

from .schema import PlanSchema, ExecutionSchema
from .targets import find_targets, filter_targets
from .api import session
from .user_input import template
from .commands import available_pull_commands
from . import git


WORKHORSE_PREFIX = os.environ.get("WORKHORSE_PREFIX", "WH-")
WORKHORSE_DIR = os.environ.get("WORKHORSE_DIR", "workhorse")


def find(name, subdir, extension):
    searches = [
        name,
        os.path.join(WORKHORSE_DIR, subdir, name),
        os.path.join(WORKHORSE_DIR, subdir, name + extension),
    ]

    for s in searches:
        if os.path.exists(s) and os.path.isfile(s):
            return s


@click.group()
def cli():
    pass


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def plan(name, token):
    """Create and save a plan"""

    session.set_token(token)

    filename = find(name, "plans", ".yml")
    if not filename:
        click.secho(f'Plan "{name}" not found', fg="red")
        exit(1)

    with open(filename, "r") as f:
        data = yaml.safe_load(f)

    click.echo(f"Loading plan at {filename}")

    p = PlanSchema().load(data)

    if "pulls" in p:
        query = p["pulls"]["search"]
        if "is:pr" not in query:
            query += " is:pr"
        search_type = "issues"
    elif "repos" in p:
        query = p["repos"]["search"]
        search_type = "repositories"

    click.echo(f'Searching GitHub for "{query}"')
    targets = find_targets(query, search_type)

    for target in targets:
        target.update_from_api()

    targets = filter_targets(targets, p["pulls"]["filter"])

    click.echo(f"{len(targets)} matching search and filter")

    print("")
    for t in targets:
        output = template.render(p["pulls"]["markdown"], t.data)
        print(output)
    print("")

    if len(targets) < 1:
        click.secho("No targets found", fg="green")
        return

    execution = ExecutionSchema().load(
        {
            "created_from": os.path.relpath(filename, os.getcwd()),
            "plan": p,
            "targets": [target.url for target in targets],
        }
    )

    execs_dir = os.path.join(WORKHORSE_DIR, "execs")
    if not os.path.exists(execs_dir):
        os.makedirs(execs_dir)

    latest = 0
    for existing in os.listdir(execs_dir):
        numbers = re.search("\d+", existing)
        if not numbers:
            continue
        latest = max(latest, int(numbers[0]))

    exec_number = latest + 1
    exec_filename = os.path.join(execs_dir, f"{WORKHORSE_PREFIX}{exec_number}.json")
    with open(exec_filename, "w+") as f:
        json.dump(execution, f, indent=2, sort_keys=True)

    click.secho("Saved for future execution!", fg="green")
    click.echo(exec_filename)


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def execute(name, token):

    session.set_token(token)

    filename = find(name, "execs", ".json")
    if not filename:
        click.secho(f'Execution "{name}" not found', fg="red")
        exit(1)

    with open(filename, "r") as f:
        data = yaml.safe_load(f)

    execution = ExecutionSchema().load(data)

    for target_url in execution["targets"]:
        click.secho(target_url, bold=True)

        for step in execution["plan"].get("pulls", {}).get("steps", []):
            for step_name, step_data in step.items():
                retry = step_data.pop("retry", False)
                # if retry True, automated
                # if retry number, retry that many times w/ auto backoff
                # if retry list of numbers, that is the backoff

                allow_error = step_data.pop("allow_error", False)
                # str to check Exception str against - if contains, let it go
                try:
                    print(f"  - {step_name} with {step_data}")
                    result = available_pull_commands[step_name](target_url, **step_data)
                except Exception as e:
                    click.secho(str(e), fg="red")
                    if allow_error and allow_error in str(e):
                        click.secho('Error allowed "{allow_error}"', fg="green")
                    # elif retry:
                    else:
                        raise e

        print("")


@cli.group()
def ci():
    pass


@ci.command()
@click.option("--force", default=False)
def plan(force):
    if git.is_dirty():
        click.secho("Git repo cannot be dirty", fg="red")

    # create plan

    if not targets:
        print("No targets found")
        # if pr exists already, close it

    # create branch, delete it if already exists and create fresh
    # commit, push

    # open PR

    # checkout -


@ci.command()
def execute():
    # func LastCommitFilesAdded(filterPrefix string) []string {
    #     cmd := exec.Command("git", "diff", "HEAD^", "HEAD", "--name-only", "--diff-filter", "A")
    #     out, err := cmd.CombinedOutput()
    #     if err != nil {
    #         panic(err)
    #     }
    #     s := string(out)

    #     lines := strings.Split(s, "\n")

    #     paths := []string{}
    #     for _, line := range lines {
    #         line := strings.TrimSpace(line)
    #         if strings.HasPrefix(line, filterPrefix) {
    #             paths = append(paths, line)
    #         }
    #     }
    #     return paths
    # }
    pass


if __name__ == "__main__":
    cli()
