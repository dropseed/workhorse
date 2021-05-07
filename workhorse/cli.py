import os

import click

from .api import session
from . import git
from .targets import Target
from .plans import Plan, load_plans
from .executions import Execution, load_executions
from .settings import (
    WORKHORSE_DIR,
    WORKHORSE_BRANCH_PREFIX,
    WORKHORSE_EXECS_DIR,
    WORKHORSE_PREFIX,
)


@click.group()
def cli():
    pass


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def prepare(name, token):
    """Prepare a plan for execution by loading current targets and saving the execution JSON"""
    session.set_token(token)

    plan = Plan.load_from_name(name)

    click.echo(f'Searching GitHub for "{plan.search}"')

    targets = plan.get_targets()

    click.echo(f"{len(targets)} matching filter")

    print("")
    for t in targets:
        output = t._render_markdown(plan.markdown)
        print(output)
    print("")

    if len(targets) < 1:
        click.secho("No targets found", fg="green")
        return (None, None)

    execution = Execution(
        created_from=os.path.relpath(plan.path, os.getcwd()),
        plan=plan,
        targets=targets,
    )

    exec_filename = execution.save()

    click.secho(f"Saved for future execution on {len(targets)} targets!", fg="green")
    click.echo(exec_filename)

    return (execution, exec_filename)


@cli.command()
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.argument("name")
def execute(name, token):
    """Execute a saved JSON"""
    session.set_token(token)

    execution = Execution.load_from_name(name)

    click.secho(f"Executing {execution.path}", bold=True, fg="green")

    execution.execute()

    click.secho(
        f"Successfully executed {name} on {len(execution.targets)} targets!", fg="green"
    )


@cli.command()
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.option("--keep", is_flag=True)
@click.argument("name")
def run(ctx, name, token, keep):
    """Prepare and execute a plan in one-step (with confirmation)"""
    session.set_token(token)

    execution, exec_filename = ctx.invoke(prepare, name=name, token=token)

    if not execution:
        return

    if not click.confirm("Execute?"):
        click.echo("Quitting")
        return

    ctx.invoke(execute, name=exec_filename, token=token)
    if not keep:
        os.remove(exec_filename)


@cli.command()
@click.pass_context
def history(ctx):
    if not os.path.exists(WORKHORSE_DIR):
        click.secho(f"{WORKHORSE_DIR} needs to exist first", fg="red")
        exit(1)

    if not os.path.exists(WORKHORSE_EXECS_DIR):
        click.secho(f"{WORKHORSE_EXECS_DIR} needs to exist first", fg="red")
        exit(1)

    plans = load_plans()
    plan_executions = {}

    for execution in load_executions():
        for plan in plans:
            if os.path.abspath(execution.created_from) == os.path.abspath(plan.path):
                plan_executions[plan] = execution
                break

        if len(plans) == len(plan_executions):
            break

    click.secho(f"{'Plan':<25} {'Execution':<10} Committed at", bold=True)
    for plan, execution in plan_executions.items():
        print(
            f"{str(plan):<25}",
            f"{str(execution):<10}",
            git.commit_datetime_of_path(execution.path) or "Unknown",
        )


@cli.group()
def ci():
    """Commands for running in GitHub Actions"""
    pass


@ci.command("prepare")
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
@click.option("--force", is_flag=True)
@click.argument("name")
def prepare_ci(ctx, name, force, token):
    """Prepare a plan for execution and create/update/close associated pull requests"""
    session.set_token(token)

    if git.is_dirty() and not force:
        click.secho("Git repo cannot be dirty", fg="red")
        exit(1)

    execution, exec_filename = ctx.invoke(prepare, name=name, token=token)

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

    if not pulls:
        response = session.post(
            f"/repos/{repo}/pulls",
            json={
                "title": title,
                "head": branch,
                "base": base,
                "body": body,
            },
        )
        response.raise_for_status()
        click.secho(f"Opened pull request: {response.json()['html_url']}")
    elif len(pulls) == 1:
        pull = pulls[0]
        if title != pull["title"] or body != pull["body"]:
            response = session.patch(
                pull["url"],
                json={
                    "title": title,
                    "body": body,
                },
            )
            response.raise_for_status()
            click.secho(f"Updated pull request: {response.json()['html_url']}")
        else:
            click.secho(f"No change to pull request: {pull['html_url']}")

    git.checkout("-")


# TODO optionally parse from git -- still have option to run a named exec
@ci.command("execute")
@click.pass_context
@click.option("--token", envvar="GITHUB_TOKEN", required=True)
def execute_ci(ctx, token):
    """Execute plans from the latest git commit"""
    session.set_token(token)

    if not git.last_commit_message().startswith(WORKHORSE_PREFIX):
        print(
            f"Not executing because commit message didn't start with {WORKHORSE_PREFIX}"
        )
        return

    added_files = git.last_commit_files_added()
    committed_execs = [
        x for x in added_files if x.startswith(WORKHORSE_EXECS_DIR + "/")
    ]
    if not committed_execs:
        print("Not executing because no exec files added by commit")
        return

    for exec_filename in committed_execs:
        ctx.invoke(execute, name=exec_filename, token=token)

    click.secho(f"Ran {len(committed_execs)} executions!", fg="green")


if __name__ == "__main__":
    cli()
