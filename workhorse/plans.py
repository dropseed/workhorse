import yaml
import os

import marshmallow

from .schema import PlanSchema
from .targets import Target
from .files import find_file
from .api import session
from .settings import WORKHORSE_DIR


def load_plans():
    plans = []

    for f in os.listdir(WORKHORSE_DIR):
        _, ext = os.path.splitext(os.path.basename(f))
        if ext == ".yml":
            try:
                plan = Plan.load_from_path(os.path.join(WORKHORSE_DIR, f))
                plans.append(plan)
            except marshmallow.ValidationError as e:
                print(e)

    return plans


class Plan:
    def __init__(self, type, search, limit, filter, markdown, steps, name, path=""):
        self.type = type
        self.search = search
        self.limit = limit
        self.filter = filter
        self.markdown = markdown
        self.steps = steps
        self.name = name
        self.path = path

    def __str__(self):
        return self.name

    def __repr__(self):
        return f"<Plan: {self}>"

    @classmethod
    def load_from_name(cls, name):
        path = find_file(name, ".yml")
        if not path:
            raise FileNotFoundError('Plan named "{name}" not found')
        return cls.load_from_path(path)

    @classmethod
    def load_from_path(cls, path):
        with open(path, "r") as f:
            data = yaml.safe_load(f)

        name, _ = os.path.splitext(os.path.basename(path))
        plan_data = PlanSchema().load(data)

        return cls(
            type=plan_data["type"],
            search=plan_data["search"],
            limit=plan_data["limit"],
            filter=plan_data["filter"],
            markdown=plan_data["markdown"],
            steps=plan_data["steps"],
            name=name,
            path=path,
        )

    def build_search_query(self):
        query = self.search

        if self.type == "pulls" and "is:pr" not in query:
            query += " is:pr"

        return query

    def find_target_urls(self):
        search_type = "issues" if type == "pulls" else "repositories"
        response = session.get(
            f"/search/{search_type}",
            params={"q": self.build_search_query(), "sort": "created", "order": "desc"},
            paginate="items",
        )
        response.raise_for_status()
        return [x["html_url"] for x in response.paginated_data]

    def get_targets(self):
        targets = []

        # TODO if no filter, then don't need to process these one by one w/ load?
        # just get the urls and limit them
        # (currently fails if filter left blank)

        for target_url in self.find_target_urls():
            target = Target(self.type, target_url)
            target._load()

            if target._expression_result(self.filter):
                targets.append(target)

            if self.limit > -1 and len(targets) >= self.limit:
                break

        return targets
