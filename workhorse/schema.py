import inspect

from marshmallow import Schema, fields, ValidationError

from .targets import Target

# def validate_pull_commands(d):
#     _validate_dict_commands(d, Pull("").get_commands())


# def validate_repo_commands(d):
#     _validate_dict_commands(d, Repo("").get_commands())


def validate_commands(d):
    commands = Target("", "")._get_commands()

    for name, data in d.items():
        if name not in commands:
            raise ValidationError(f"{name} is not an available command")

        func = commands[name]

        input_params = list(data.keys())

        available_params = list(inspect.signature(func).parameters.keys())
        available_params.append(
            "retry"
        )  # retry is always available, but actually outside the function itself
        available_params.append(
            "allow_error"
        )  # retry is always available, but actually outside the function itself

        if not set(input_params).issubset(set(available_params)):
            raise ValidationError(
                f"Options for {name} don't match what is available:\n\n{input_params}\n\n{available_params}"
            )


class PlanSchema(Schema):
    type = fields.Str(required=True)  # repos, pulls, issues
    search = fields.Str(required=True)
    markdown = fields.Str(required=True)
    filter = fields.Str()
    steps = fields.List(
        fields.Dict(
            keys=fields.Str(), values=fields.Dict(), validate=validate_commands
        )
    )
    limit = fields.Int(default=-1)


# class PlanSchema(Schema):
#     # TODO validate only one of these
#     pulls = fields.Nested(PullsSchema())
#     repos = fields.Nested(ReposSchema())


class ExecutionSchema(Schema):
    created_from = fields.Str()
    plan = fields.Nested(PlanSchema(), required=True)
    targets = fields.List(fields.Str())
