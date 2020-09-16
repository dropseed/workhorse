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


def get_api_url(url):
    """Get a API URL (without base) for a given HTML URL or API URL"""
    match = re.search(r"/([^/]+)/([^/]+)/(pull|issue)/(\d+)$", url)
    return f"repos/{match[1]}/{match[2]}/{match[3]}s/{match[4]}"


# def get_repo_api_url(url):
#     """Get a API URL (without base) for a given HTML URL or API URL"""
#     match = re.search(r"/([^/]+)/([^/]+)/", url)
#     return f"repos/{match[1]}/{match[2]}"


class Target:
    def __init__(self, url):
        self.url = url
        self.data = {}

        self.api_url = get_api_url(self.url)

    def update_from_api(self):
        response = session.get(self.api_url)
        response.raise_for_status()
        self.data = response.json()
