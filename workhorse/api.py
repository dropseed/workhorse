from requests import Session
from urllib.parse import urljoin


class APISession(Session):
    def __init__(self, base_url=None, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.base_url = base_url
        self.params.update({"per_page": 100})

    def request(self, method, url, *args, **kwargs):
        next_url = urljoin(self.base_url, url)
        paginate = kwargs.pop("paginate", False)
        paginated_data = []

        while next_url:
            response = super().request(method, next_url, *args, **kwargs)
            print(f"{method} {response.status_code} {next_url}")

            if response.text:
                if isinstance(paginate, str):
                    paginated_data += response.json().get(paginate, [])
                else:
                    paginated_data += response.json()

            if paginate:
                next_url = response.links.get("next", {}).get("url", None)
            else:
                next_url = None

        # custom property for paginated, combined data
        response.paginated_data = paginated_data

        return response

    def set_token(self, token):
        self.headers.update({"Authorization": f"token {token}"})


session = APISession(base_url="https://api.github.com")
