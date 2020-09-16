import re

from .api import session
from .user_input.expressions import Expression


def find_targets(query, search_type):
    response = session.get(
        f"/search/{search_type}",
        params={"q": query, "sort": "created", "order": "desc"},
        paginate="items",
    )
    response.raise_for_status()
    targets = [Target(x["html_url"]) for x in response.paginated_data]
    return targets


def filter_targets(targets, filter):
    return [target for target in targets if Expression(filter, target.data).compile()]


class Target:
    def __init__(self, url):
        self.url = url
        self.data = {}

        match = re.search(r"/([^/]+)/([^/]+)/(pull|issue)/(\d+)$", self.url)
        self.api_url = f"repos/{match[1]}/{match[2]}/{match[3]}s/{match[4]}"

    def update_from_api(self):
        response = session.get(self.api_url)
        response.raise_for_status()
        self.data = response.json()
