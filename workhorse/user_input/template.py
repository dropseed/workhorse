# Copyright (C) Dropseed, LLC - All Rights Reserved
# Unauthorized copying of files herein, via any medium is strictly prohibited
# Proprietary and confidential
# Written by Dropseed, LLC <support@dropseed.io>, 2018
from typing import Any, Dict

from jinja2 import StrictUndefined
from jinja2.exceptions import UndefinedError
from jinja2.sandbox import ImmutableSandboxedEnvironment


def render(template: str, context: Dict[str, Any] = {}) -> str:
    environment = ImmutableSandboxedEnvironment(undefined=StrictUndefined)
    environment.globals = context

    jinja_template = environment.from_string(template)
    output = None

    try:
        output = jinja_template.render()
    except UndefinedError as e:
        raise TemplateException(e.message)
    except TypeError as e:
        if str(e) == "no loader for this environment specified":
            raise TemplateException("Extending templates is not allowed")
        else:
            raise e

    return output


class TemplateException(Exception):
    pass
