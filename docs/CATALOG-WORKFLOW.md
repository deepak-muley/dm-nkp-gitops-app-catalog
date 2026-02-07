# Catalog Workflow Script

`./catalog-workflow.sh` orchestrates add-application, validate, add-tests, and build-push in one script.

## Commands

| Command | Description | Required Options |
|---------|-------------|------------------|
| `add-app` | Add application via add-application.sh | `--appname`, `--version`, `--ocirepo` |
| `validate` | Run validate.sh (nkp validate + ghcr.io login) | — |
| `add-tests` | Create apptest placeholders for Ginkgo tests | `--appname <name>` or `--all` |
| `build-push` | Build and push catalog bundle | `--tag <version>` |
| `all` | Run validate + build-push | `--tag <version>` |

## Examples

```bash
# Add one app (Helm)
./catalog-workflow.sh add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo

# Add Kubeflow component (Kustomize)
./catalog-workflow.sh add-app --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default --ref release-v1.10

# Validate
./catalog-workflow.sh validate

# Create test placeholders for one app
./catalog-workflow.sh add-tests --appname podinfo

# Create test placeholders for all apps
./catalog-workflow.sh add-tests --all

# Build and push
./catalog-workflow.sh build-push --tag v0.1.0

# Validate + build-push in one go
./catalog-workflow.sh all --tag v0.1.0

# Full flow: add app, validate, add tests, build
./catalog-workflow.sh add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo --force
./catalog-workflow.sh validate
./catalog-workflow.sh add-tests --appname podinfo
./catalog-workflow.sh build-push --tag v0.1.0
```

## Batch Add (multiple apps)

Use a loop with commands from [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md):

```bash
# Example: add multiple apps from the commands doc
apps=(
  "podinfo:6.9.4:oci://ghcr.io/stefanprodan/charts/podinfo"
  "cert-manager:1.19.2:oci://quay.io/jetstack/charts/cert-manager"
  "traefik:38.0.2:oci://ghcr.io/traefik/helm/traefik"
)
for entry in "${apps[@]}"; do
  IFS=: read -r name version ocirepo <<< "$entry"
  ./catalog-workflow.sh add-app --appname "$name" --version "$version" --ocirepo "$ocirepo" --force
done
./catalog-workflow.sh validate
```

## add-tests Placeholders

`add-tests` creates:
- `apptests/appscenarios/<app>.go` — AppScenario stub (go:build ignore)
- `apptests/suites/<app>_test.go` — Test suite stub
- `apptests/README.md` — Setup instructions (on first run)

These are placeholders. See [APP-TESTS-GUIDE.md](APP-TESTS-GUIDE.md) for full Ginkgo/Kind setup.

## justfile (Apptests)

From repo root, use [just](https://just.systems/) to run apptests:

```bash
just apptests              # run all tests
just apptests-app podinfo  # run tests for one app
just apptests-install      # install-label tests only
just apptests-upgrade      # upgrade-label tests only
just apptests-tidy         # go mod tidy
```
