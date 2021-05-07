import re
import os
import json

import click
import marshmallow

from .schema import ExecutionSchema
from .settings import WORKHORSE_EXECS_DIR, WORKHORSE_PREFIX
from .targets import Target
from .files import find_file


def load_executions():
    plans = []

    for f in os.listdir(WORKHORSE_EXECS_DIR):
        _, ext = os.path.splitext(os.path.basename(f))
        if ext == ".json":
            try:
                plan = Execution.load_from_path(os.path.join(WORKHORSE_EXECS_DIR, f))
                plans.append(plan)
            except marshmallow.ValidationError as e:
                print(e)

    return plans


class Execution:
    def __init__(self, created_from, plan, targets, name, path=""):
        self.created_from = created_from
        self.plan = plan
        self.targets = targets
        self.name = name
        self.path = path

    def __str__(self):
        return self.name

    def __repr__(self):
        return f"<Execution: {self}>"

    @classmethod
    def load_from_name(cls, name):
        path = find_file(name, ".json", specific_dir=WORKHORSE_EXECS_DIR)
        if not path:
            raise FileNotFoundError('Execution named "{name}" not found')
        return cls.load_from_path(path)

    @classmethod
    def load_from_path(cls, path):
        with open(path, "r") as f:
            data = json.load(f)

        name, _ = os.path.splitext(os.path.basename(path))
        execution_data = ExecutionSchema().load(data)

        return cls(
            created_from=execution_data["created_from"],
            plan=execution_data["plan"],
            targets=execution_data["targets"],
            name=name,
            path=path,
        )

    def dump(self):
        return ExecutionSchema.dump(
            {
                "created_from": self.created_from,
                "plan": self.plan,
                "targets": [target._url for target in self.targets],
            }
        )

    def save(self):
        if not os.path.exists(WORKHORSE_EXECS_DIR):
            os.makedirs(WORKHORSE_EXECS_DIR)

        latest = 0
        for existing in os.listdir(WORKHORSE_EXECS_DIR):
            numbers = re.search("\d+", existing)
            if not numbers:
                continue
            latest = max(latest, int(numbers[0]))

        exec_number = latest + 1
        exec_filename = os.path.join(
            WORKHORSE_EXECS_DIR, f"{WORKHORSE_PREFIX}{exec_number}.json"
        )
        with open(exec_filename, "w+") as f:
            json.dump(self.dump(), f, indent=2, sort_keys=True)

        return exec_filename

    def execute(self):
        for target_url in self.targets:
            click.secho(target_url, bold=True, fg="cyan")

            target = Target(self.plan.type, target_url)
            target._load()

            for step in self.plan.steps:
                target.execute_step(step)
                print("")

            print("")
