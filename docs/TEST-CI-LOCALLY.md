# Testing the CI Workflow Locally

You can mirror what CI does on every push in two ways.

## Option 1: catalog-workflow.sh (recommended)

Run the same sequence CI runs (validate then apptests) via the main workflow script:

```bash
./catalog-workflow.sh ci-local
```

Optional: limit apptests to one app or a label:

```bash
./catalog-workflow.sh ci-local --appname podinfo
./catalog-workflow.sh ci-local --label install
```

This runs:

1. **validate** — catalog structure, YAML, metadata, and `nkp validate` (if `nkp` is available)
2. **apptests** — Ginkgo/Kind integration tests in `apptests/` (via `just apptests` if present, else `go test` directly)

Requires: Go, Docker (for apptests). Optional: `nkp`, `yq`, and [just](https://just.systems/) (used for apptests when available).

### Via just

```bash
just ci-local   # calls ./catalog-workflow.sh ci-local
```

### Step by step (without ci-local)

```bash
./catalog-workflow.sh validate
./catalog-workflow.sh test
# Or for one app: ./catalog-workflow.sh test --appname podinfo
```

If you don’t have `nkp`, the workflow’s validate job still does structure + YAML + metadata checks in CI; locally `./catalog-workflow.sh validate` will do what it can without `nkp`.

---

## Option 2: Run the workflow file with act

[act](https://github.com/nektos/act) runs GitHub Actions workflows locally using Docker.

### Install act

- **macOS (Homebrew):** `brew install act`
- **Linux:** see [act releases](https://github.com/nektos/act/releases)

### Run the CI workflow

From the repo root:

```bash
# List available workflows and events
act -l

# Run the CI workflow (push event, default branch)
act push

# Run on a specific event
act pull_request
```

`act` uses Docker to run each job in a container, so behavior should be close to GitHub Actions. The first run may be slow while images are pulled.

### Limitations

- Secrets (e.g. `GITHUB_TOKEN`) are mocked unless you pass `--secret-file` or `-s`.
- The workflow’s **build** job only does a dry-run locally; actual bundle build/push is intended for CI or manual runs.
- Jobs that need `nkp` will fail in act unless you install it in the job or use a custom image.

---

## Summary

| Goal                    | Command / approach                                      |
|-------------------------|---------------------------------------------------------|
| Run CI locally          | `./catalog-workflow.sh ci-local`                       |
| CI local (one app)      | `./catalog-workflow.sh ci-local --appname podinfo`      |
| Via just                | `just ci-local` (delegates to catalog-workflow.sh)      |
| Validate only           | `./catalog-workflow.sh validate`                       |
| Apptests only           | `./catalog-workflow.sh test` or `just apptests`         |
| Run the workflow as-is  | `act push` (with act installed)                        |

See also: [CATALOG-WORKFLOW.md](CATALOG-WORKFLOW.md), [APP-TESTS-GUIDE.md](APP-TESTS-GUIDE.md).
