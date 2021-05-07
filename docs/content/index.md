
# Manage GitHub PRs and Repos in Bulk

Workhorse helps make changes across your GitHub organization by executing **plans**.

You can write plans for one-off situations (ex. finding and replacing a string across repos),
or you can set up recurring plans that run via GitHub Actions (ex. to merge dependency update PRs once a week).

Plans are written in YAML, using API abstractions that make it easier to do things like:

- find and replace a string across repos
- perform org-wide migrations
- merge dependency PRs across the organization
- clone repos and run a shell script across them

## How it works

![Workhorse overview diagram](/assets/img/workhorse-overview.png)

### Searching

Workhorse uses GitHub's built-in search queries to get the initial candidates to run your plan on.

You can search for repos:

```yaml
type: repos
search: "archived:false org:yourorg"
```

Or pull requests:

```yaml
type: pulls
search: "archived:false org:yourorg is:open label:dependencies"
```

### Filtering

But you usually need to be more specific than the search API can be.

If you're going to do a find and replace, for example, then you want to filter the results down to only repos where that file exists and where it still has the old string.

```yaml
type: repos
search: "archived:false org:yourorg"
filter: '"192.168.1.XXX" in repo.file_contents("config.json")'
```

### Running "steps"

Your search+filter will return a set of repos (or pull requests).
On each one of those items,
you'll run the same series of operations.
This is where the "write once, run X times" comes in.

For common tasks that take a bunch of steps or API calls,
we have a specific **step** to do the heavy-lifting.
Like this one for `replace_in_file`:

```yaml
type: repos
search: "archived:false org:yourorg"
filter: '"192.168.1.XXX" in repo.file_contents("config.json")'
steps:
- replace_in_file:
    file: config.json
    find: "192.168.1.XXX"
    replace: "192.168.1.YYY"
    branch: master
    message: "Change IP address in config.json"
```

But you aren't limited to the specific features that we "support", because you can always drop back to the shell (with optional access to a full git clone):

```yaml
steps:
- clone: {}
- shell:
    run: git checkout -b github-actions
- shell:
    run: |
        rm -r .circleci
        mkdir -p .github/workflows
...
```

Or raw GitHub API calls:

```yaml
steps:
- api:
    method: put
    repo_url: /branches/master/protection
    json:
      required_status_checks:
        strict: false
        contexts: [ci]
        enforce_admins: null
        required_pull_request_reviews: null
        restrictions: null
```

### Execute plans as pull requests

You can run Workhorse directly on your command line,
but recurring plans can also run via GitHub Actions.

Through GitHub Actions, each **execution** will be presented as a PR:

![Workhorse pull request example](/assets/img/workhorse-pr-example.png)

Merging the PR will run the plan exactly as you see it -- the targets (repos or pulls) are locked to what is committed in the PR.
