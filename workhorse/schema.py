import inspect

from marshmallow import Schema, fields, ValidationError

from .commands import available_pull_commands, available_repo_commands


def validate_pull_commands(d):
    _validate_dict_commands(d, available_pull_commands)


def validate_repo_commands(d):
    _validate_dict_commands(d, available_repo_commands)


def _validate_dict_commands(d, commands):
    for name, data in d.items():
        if name not in commands:
            raise ValidationError(f"{name} is not an available command")

        func = commands[name]
        input_params = data.keys()
        available_params = inspect.signature(func).parameters.keys()
        if not set(input_params).issubset(set(available_params)):
            raise ValidationError(f"Options for {name} don't match what is available:\n\n{input_params}\n\n{available_params}")


class PullsSchema(Schema):
    search = fields.Str(required=True)
    filter = fields.Str()
    markdown = fields.Str(default="{{ title }}")
    steps = fields.List(fields.Dict(keys=fields.Str(), values=fields.Dict(), validate=validate_pull_commands))


class ReposSchema(Schema):
    search = fields.Str(required=True)
    filter = fields.Str()
    markdown = fields.Str(default="{{ full_name }}")
    steps = fields.List(fields.Dict(keys=fields.Str(), values=fields.Dict(), validate=validate_repo_commands))


class PlanSchema(Schema):
    # TODO validate only one of these
    pulls = fields.Nested(PullsSchema())
    repos = fields.Nested(ReposSchema())


class ExecutionSchema(Schema):
    created_from = fields.Str()
    plan = fields.Nested(PlanSchema(), required=True)
    targets = fields.List(fields.Str())
