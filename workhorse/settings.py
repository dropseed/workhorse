import os


WORKHORSE_PREFIX = os.environ.get("WORKHORSE_PREFIX", "WH-")
WORKHORSE_DIR = os.environ.get("WORKHORSE_DIR", "workhorse")
WORKHORSE_BRANCH_PREFIX = os.environ.get("WORKHORSE_BRANCH_PREFIX", "workhorse/")
WORKHORSE_EXECS_DIR = os.environ.get(
    "WORKHORSE_EXECS_DIR", os.path.join(WORKHORSE_DIR, "execs")
)
