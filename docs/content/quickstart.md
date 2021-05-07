# Quickstart

## Trying it out

Workhorse is a Python package that can be installed with pip.

Use your package manager of choice, or pip itself:

```console
$ pip3 install -U workhorse
```

```yaml
type: pulls
search: "is:open is:pr label:dependencies org:yourorg archived:false"
markdown: "- {{ pull.html_url }}\n  {{ pull.title }}"
filter: 'pull.mergeable_state == "clean" and pull.mergeable'
steps:
- merge:
    merge_method: squash
```

## Creating a repo for recurring plans

Typically you'll want a dedicated repo for storing your Workhorse plans and executing them via GitHub Actions.

Copy the workhorse-template repo as a starting point.

```console
$ git clone https://github.com/yourorg/workhorse
$ ./scripts/install
$ ./scripts/workhorse new
```

TODO workhorse-template repo

##
