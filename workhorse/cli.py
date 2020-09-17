import time
import os
import json
import re
import requests
import random

import yaml
import click

from .schema import PlanSchema, ExecutionSchema
from .api import session
from . import git
from .targets import Target
from .exceptions import RetryException


WORKHORSE_PREFIX = os.environ.get("WORKHORSE_PREFIX", "WH-")
WORKHORSE_DIR = os.environ.get("WORKHORSE_DIR", "workhorse")
WORKHORSE_BRANCH_PREFIX = os.environ.get("WORKHORSE_BRANCH_PREFIX", "workhorse/")


def find(name, extension, subdir=""):
    searches = [
        name,
        os.path.join(WORKHORSE_DIR, subdir, name),
        os.path.join(WORKHORSE_DIR, subdir, name + extension),
    ]

    for s in searches:
        if os.path.exists(s) and os.path.isfile(s):
            return s


def find_target_urls(query, type, search_type):
    response = session.get(
        f"/search/{search_type}",
        params={"q": query, "sort": "created", "order": "desc"},
        paginate="items",
    )
    response.raise_for_status()
    return [x["html_url"] for x in response.paginated_data]


@click.group()
def cli():
    pass


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def plan(name, token):
    """Create and save a plan"""
    session.set_token(token)

    filename = find(name, ".yml")
    if not filename:
        click.secho(f'Plan "{name}" not found', fg="red")
        exit(1)

    with open(filename, "r") as f:
        data = yaml.safe_load(f)

    click.echo(f"Loading plan at {filename}")

    p = PlanSchema().dump(PlanSchema().load(data))
    type = p["type"]
    query = p["search"]

    if type == "pulls":
        if "is:pr" not in query:
            query += " is:pr"
        search_type = "issues"

    elif type == "repos":
        search_type = "repositories"

    click.echo(f'Searching GitHub for "{query}"')

    limit = p["limit"]
    targets = []
    for target_url in find_target_urls(query, type, search_type):
        target = Target(type, target_url)
        target._load()

        if target._expression_result(p["filter"]):
            targets.append(target)

        if limit > -1 and len(targets) >= limit:
            click.secho(f"Limiting to {limit}", fg="yellow")
            break

    click.echo(f"{len(targets)} matching filter")

    print("")
    for t in targets:
        output = t._render_markdown(p["markdown"])
        print(output)
    print("")

    if len(targets) < 1:
        click.secho("No targets found", fg="green")
        return (None, None)

    execution = ExecutionSchema().load(
        {
            "created_from": os.path.relpath(filename, os.getcwd()),
            "plan": p,
            "targets": [target._url for target in targets],
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

    click.secho(f"Saved for future execution on {len(targets)} targets!", fg="green")
    click.echo(exec_filename)

    return (execution, exec_filename)


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def execute(name, token):
    session.set_token(token)

    filename = find(name, ".json", "execs")
    if not filename:
        click.secho(f'Execution "{name}" not found', fg="red")
        exit(1)

    with open(filename, "r") as f:
        data = yaml.safe_load(f)

    execution = ExecutionSchema().load(data)
    p = execution["plan"]

    targets = execution["targets"]
    for target_url in targets:
        # enumerate and show 2/13 for progress?
        click.secho(target_url, bold=True, fg="cyan")
        target = Target(p["type"], target_url)
        target._load()

        for step in p.get("steps", []):
            for step_name, step_data in step.items():
                retry = step_data.pop("retry", False)
                # if retry True, automated
                # if retry number, retry that many times w/ auto backoff
                # if retry list of numbers, that is the backoff

                allow_error = step_data.pop("allow_error", False)

                attempt = 0
                while True:
                    attempt = attempt + 1

                    try:
                        # enumerate and show 2/13 for progress?
                        click.secho(f"- {step_name}", bold=True)
                        for k, v in step_data.items():
                            click.secho(f"    {k}: {str(v)[:70]}", bold=True)
                        result = target._run_command(step_name, step_data)
                        break

                    except Exception as e:
                        if allow_error and (
                            allow_error is True or allow_error in str(e)
                        ):
                            click.secho('Error allowed "{allow_error}"', fg="green")
                            break

                        click.secho(str(e), fg="red")

                        if isinstance(e, requests.RequestException):
                            click.secho(e.response.text, fg="red")

                        if isinstance(e, (requests.RequestException, RetryException)):
                            if retry and isinstance(retry, list):
                                backoff_index = attempt - 1
                                if backoff_index < len(retry):
                                    backoff = retry[attempt - 1]
                                    click.secho(
                                        f"Waiting {backoff} seconds to retry...",
                                        fg="yellow",
                                    )
                                    time.sleep(backoff)
                                    continue

                            elif retry and isinstance(retry, int):
                                if attempt <= retry:
                                    click.secho(
                                        "Waiting 5 seconds to retry...", fg="yellow"
                                    )
                                    time.sleep(5)
                                    continue

                        raise e

            print("")
        print("")

    click.secho(f"Successfully executed {name} on {len(targets)} targets!", fg="green")


@cli.command()
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.option("--keep", is_flag=True)
@click.argument("name")
def run(ctx, name, token, keep):
    """Plan and execute in one go"""
    session.set_token(token)

    confirm = random.choice(["yes", "yep", "ok", "yeah"])
    if (
        click.prompt(
            f"Are you sure you want to run {name}? This could be destructive. Enter '{confirm}' to continue"
        )
        != confirm
    ):
        click.echo("Quitting")
        return

    execution, exec_filename = ctx.invoke(plan, name=name, token=token)

    if not click.confirm("Execute?"):
        click.echo("Quitting")
        return

    ctx.invoke(execute, name=exec_filename, token=token)
    if not keep:
        os.remove(exec_filename)


@cli.group()
def ci():
    pass


@ci.command("plan")
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.option("--force", is_flag=True)
@click.argument("name")
def plan_ci(ctx, name, force, token):
    session.set_token(token)

    if git.is_dirty() and not force:
        click.secho("Git repo cannot be dirty", fg="red")
        exit(1)

    execution, exec_filename = ctx.invoke(plan, name=name, token=token)

    plan_name = os.path.splitext(os.path.basename(name))[0]

    base = "master"
    branch = f"{WORKHORSE_BRANCH_PREFIX}{plan_name}"

    repo = git.repo_from_remote()

    head = f"{repo.split('/')[0]}:{branch}"
    response = session.get(
        f"/repos/{repo}/pulls", params={"state": "open", "base": base, "head": head}
    )
    response.raise_for_status()
    pulls = response.json()
    for pull in pulls:
        click.secho(
            f"Found an existing pull request for this plan: {pull['html_url']}",
            fg="yellow",
        )

    if not execution:
        if len(pulls) == 1:
            pull = pulls[0]

            response = session.patch(pull["url"], json={"state": "closed"})
            response.raise_for_status()

            response = session.delete(
                pull["head"]["repo"]["git_refs_url"].replace(
                    "{/sha}", f"/heads/{branch}"
                )
            )
            response.raise_for_status()
        return

    try:
        git.create_branch(branch)
    except Exception:
        click.secho("Branch already exists, deleting it", fg="yellow")
        # TODO how to only push if changes were made?
        # stash, checkout branch, apply, see if diff?
        git.delete_branch(branch)
        git.create_branch(branch)

    exec_name = os.path.splitext(os.path.basename(exec_filename))[0]
    title = f"{exec_name}: {plan_name}"
    git.add_commit(exec_filename, title)
    git.push(branch)

    body = f"Merging this PR will run {plan_name} on the following PRs:\n\n"
    for url in execution["targets"]:
        target = Target(execution["plan"]["type"], url)
        target._load()
        md = target._render_markdown(execution["plan"]["markdown"])
        body = body + "\n" + md

    if len(pulls) == 1:
        response = session.post(
            f"/repos/{repo}/pulls",
            json={
                "title": title,
                "head": branch,
                "base": base,
                "body": body,
            },
        )
    else:
        response = session.patch(
            f"/repos/{repo}/pulls",
            json={
                "title": title,
                "body": body,
            },
        )
    response.raise_for_status()

    click.secho(f"Opened pull request: {response.json()['html_url']}")

    git.checkout("-")


@ci.command("execute")
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
def execute_ci(ctx, token):
    session.set_token(token)

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
