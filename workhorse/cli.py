import os
import json

import yaml
import click

from .schema import PlanSchema, ExecutionSchema
from .targets import find_targets, filter_targets, Target
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

    p = PlanSchema().load(data)
    print(p)

    # custom validation?

    if "pulls" in p:
        query = p["pulls"]["search"]
        if "is:pr" not in query:
            query += " is:pr"
        search_type = "issues"
    elif "repos" in p:
        query = p["repos"]["search"]
        search_type = "repositories"

    targets = find_targets(query, search_type)

    for target in targets:
        target.update_from_api()

    targets = filter_targets(targets, p["pulls"]["filter"])

    for t in targets:
        output = template.render(p["pulls"]["markdown"], t.data)
        print(output)

    if len(targets) < 1:
        click.secho("No targets found", fg="green")
        return

    execution = ExecutionSchema().load({
        "created_from": filename,
        "plan": p,
        "targets": [target.url for target in targets],
    })

    execs_dir = os.path.join(WORKHORSE_DIR, "execs")
    if not os.path.exists(execs_dir):
        os.makedirs(execs_dir)

    latest = 0
    for existing in os.listdir(execs_dir):
        numbers = re.search("\d+", existing)
        if not numbers:
            continue
        latest = max(latest, numbers[0])

    exec_number = latest + 1

    with open(os.path.join(execs_dir, f"{WORKHORSE_PREFIX}{exec_number}.yml"), "w+") as f:
        json.dump(execution, f, indent=2, sort_keys=True)


@cli.command()
@click.argument("name")
def execute(name):
    filename = find(name, "execs", ".json")
    if not filename:
        click.secho(f'Execution "{name}" not found', fg="red")
        exit(1)

    with open(filename, "r") as f:
        data = yaml.safe_load(f)

    execution = ExecutionSchema().load(data)

    for target_url in execution["targets"]:
        print(target_url)
        target = Target(target_url)

        for step in execution["plan"].get("pulls", {}).get("steps", []):
            for step_name, step_data in step.items():
                # if func receives target_url (how to tell?) then pass it too
                result = available_pull_commands[step_name](**step_data)


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
