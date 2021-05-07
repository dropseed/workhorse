import os

from .settings import WORKHORSE_DIR


def find_file(name, extension, specific_dir=""):
    searches = [
        name,
        os.path.join(WORKHORSE_DIR, name),
    ]

    if specific_dir:
        searches.append(os.path.join(specific_dir, name + extension))

    for s in searches:
        if os.path.exists(s) and os.path.isfile(s):
            return s
