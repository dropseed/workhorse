import os
import yaml
import requests
import sys
import time


api_base = "https://api.github.com"


def run(path, token):
    with open(path, "r") as f:
        data = yaml.safe_load(f)

    actions = {"add_labels": add_labels, "remove_label": remove_label, "sleep": sleep}

    session = requests.session()
    session.headers.update({"Authorization": f"token {token}"})

    search = data["search"]

    search_results = []
    next_page_url = api_base + "/search/issues"
    while next_page_url:
        response = session.get(next_page_url, params=search["issues"])
        response.raise_for_status()
        next_page_url = response.links.get("next", "")
        search_results += response.json()["items"]

    print(f"{len(search_results)} search results")

    for item in search_results:

        repo_full_name = item["repository_url"].split("https://api.github.com/repos/")[
            1
        ]
        issue_number = item["number"]
        html_url = item["html_url"]
        print(html_url)

        for step in data["steps"]:
            # assume each step is a dict...
            for action_name, action_arg in step.items():
                print(f"Running {action_name}: {action_arg}")
                actions[action_name](
                    action_arg,
                    repo_full_name=repo_full_name,
                    issue_number=issue_number,
                    session=session,
                )

        print()


# these get "registered", and there is a system for registering outside of built-ins from me
# each should maybe be a class? with a validate fn, that can run as a part of config loading/validating to make sure args
# will work -- if were a configyaml class, then this would work
def add_labels(labels, repo_full_name, issue_number, session):
    response = session.post(f"{api_base}/repos/{repo_full_name}/issues/{issue_number}/labels", json={"labels": labels})
    response.raise_for_status()


def remove_label(label, repo_full_name, issue_number, session):
    response = session.delete(
        f"{api_base}/repos/{repo_full_name}/issues/{issue_number}/labels/{label}")
    response.raise_for_status()


def sleep(seconds, *args, **kwargs):
    time.sleep(seconds)

if __name__ == "__main__":
    run(sys.argv[1], sys.argv[2])
