# Limit

Limit the number of pull requests or repos.

This can be used for testing a plan (i.e. `limit: 1` and test the real world results),
or it can be used for executing a plan in batches (i.e. `limit: 5` for more reviewable and/or phased rollouts of a plan).

```yaml
type: pulls
limit: 5
```
