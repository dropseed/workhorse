# workhorse (beta)

Workhorse is a tool for managing GitHub repos and pull requests in bulk.

## CI - recurring checks with PRs for resolution

For example

- Merge dependency update PRs where tests are passing
- Close outdated PRs that haven't had activity in 90 days

## CLI - one-off changes across an organization

For example:

- Update all or part of a YAML config file across similar repos (e.g. `.github/workflows/test.yml`)
- Change repo settings across an organization (e.g. disable rebase merging)
