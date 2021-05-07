# GitHub Actions

```yaml
name: workhorse

on:
  push:
    branches:
    - master
  schedule:
  - cron: "0 6 * * *"
  workflow_dispatch: {}

jobs:
  workhorse:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 2
    - uses: actions/setup-python@v2
    - run: pip install "workhorse<1.0.0"

    - run: workhorse ci execute
      if: github.event_name == 'push'
      env:
        GITHUB_TOKEN: ${{ secrets.WORKHORSE_GITHUB_TOKEN }}

    - run: |
        git config user.email "github-actions@github.com"
        git config user.name "github-actions"

    - run: workhorse ci plan merge-deps
      env:
        GITHUB_TOKEN: ${{ secrets.WORKHORSE_GITHUB_TOKEN }}
    - run: workhorse ci plan close-unmergeable-deps
      env:
        GITHUB_TOKEN: ${{ secrets.WORKHORSE_GITHUB_TOKEN }}
```
