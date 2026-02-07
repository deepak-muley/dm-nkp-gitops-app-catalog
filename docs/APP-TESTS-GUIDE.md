# App Tests Guide

This guide describes how to run and add application tests for this catalog. Tests are specific to this catalog (applications/ layout). They use **Ginkgo** BDD-style tests that:
1. Spin up a **Kind** cluster
2. Install **Flux**
3. Deploy each app via its helmrelease kustomization
4. Verify **HelmRelease** status (Ready, UpgradeSucceeded)

## Architecture

```
apptests/
├── appscenarios/       # AppScenario implementations per app
│   ├── constant/       # Shared constants (DEFAULT_NAMESPACE, POLL_INTERVAL)
│   ├── podinfo.go
│   └── traefik_hub.go
├── suites/             # Ginkgo test specs per app
│   ├── suites_test.go  # Setup: Kind, Flux, k8s client
│   ├── podinfo_test.go
│   └── traefikhub_test.go
├── utils/              # Helpers (AbsolutePathTo, GetPrevVAppsUpgradePath)
├── go.mod
└── main.go
```

## AppScenario Interface

Implement in `appscenarios/<app>.go`:

```go
type AppScenario interface {
    Name() string
    Install(ctx context.Context, env *environment.Env) error
    InstallPreviousVersion(ctx context.Context, env *environment.Env) error
    Upgrade(ctx context.Context, env *environment.Env) error
}
```

- **Name()** — Application name (e.g. `"podinfo"`).
- **Install** — Deploy latest version; use `utils.AbsolutePathTo(name, version)` for path.
- **InstallPreviousVersion** — Deploy second-latest version; use `utils.GetPrevVAppsUpgradePath(name)`.
- **Upgrade** — Deploy latest (upgrade path); use `utils.AbsolutePathTo(name, "")` for latest.

## Install Flow

1. Resolve app path: `applications/<app>/<version>/helmrelease`
2. Apply kustomization with substitution: `releaseName`, `releaseNamespace`
3. Uses `env.ApplyKustomizations(ctx, helmreleasePath, map[string]string{...})` from kommander-applications/apptests.

## Test Specs (suites/<app>_test.go)

### Install Test
- `Describe("Installing <app>", Label("install"), ...)`
- Call `pr.Install(ctx, env)`
- `Eventually` check HelmRelease has `ReadyCondition == True`

### Upgrade Test
- `Describe("Upgrading <app>", Label("upgrade"), ...)`
- **Step 1**: `pr.InstallPreviousVersion(ctx, env)` — install older version
- **Step 2**: `Eventually` verify HelmRelease Ready
- **Step 3**: `pr.Upgrade(ctx, env)` — upgrade to latest
- **Step 4**: `Eventually` verify HelmRelease Ready and `Reason == UpgradeSucceededReason`

## Setup

```bash
./setup.sh
```

Runs `go mod tidy` in apptests/ to ensure dependencies are available. Tests are specific to this catalog only; no external clone.

## Dependencies

The apptests module depends on:
- `github.com/mesosphere/kommander-applications/apptests` — Kind, Flux, environment
- `github.com/onsi/ginkgo/v2`, `github.com/onsi/gomega`
- `github.com/fluxcd/helm-controller/api`
- `sigs.k8s.io/kind`

## Adding Tests to dm-nkp-gitops-app-catalog

### Option A: Use existing apptests (recommended)

1. Run `./setup.sh` — ensures Go dependencies are available.
2. Run `just apptests` or `just apptests-app podinfo`.
3. Add `AppScenario` in `appscenarios/<app>.go` and `suites/<app>_test.go` for additional apps in this catalog.

### Option C: Lightweight validation (current approach)

- Use `./catalog-workflow.sh validate` (nkp validate catalog-repository) for manifest validation.
- Use `.github/workflows/ci.yml` for structure and YAML checks.
- No Kind/Flux integration tests.

### Option D: Gradual adoption

1. Add `apptests/` directory with `go.mod`, `utils/`, `constant/`.
2. Start with one app (e.g. podinfo) — implement AppScenario + test suite.
3. Add devbox or just for `validate-manifests` and `test-apps` commands.
4. Integrate into CI when ready.

## Running Tests

### Via justfile (repo root)

```bash
just apptests              # run all tests
just apptests-app podinfo  # run tests for one app
just apptests-install      # install-label tests only
just apptests-upgrade      # upgrade-label tests only
just apptests-tidy         # go mod tidy
```

### Directly

```bash
cd apptests
go test ./suites/ -v -run "podinfo"
# Or with labels:
go test ./suites/ -v -ginkgo.label-filter="install && podinfo"
```

## References

- [nkp-partner-catalog](https://github.com/nutanix-cloud-native/nkp-partner-catalog)
- [nkp-partner-catalog apptests](https://github.com/nutanix-cloud-native/nkp-partner-catalog/tree/main/apptests)
- [CONTRIBUTING.md — Testing applications](https://github.com/nutanix-cloud-native/nkp-partner-catalog/blob/main/CONTRIBUTING.md#testing-applications)
