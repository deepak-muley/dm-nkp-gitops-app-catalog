# App Tests

Ginkgo/Kind integration tests for NKP catalog applications. Tests spin up a Kind cluster, install Flux, deploy each app via its HelmRelease, and verify the deployment succeeds.

## What Do Apptests Do?

Apptests exercise the full deploy path for catalog applications:

1. **Provision Kind cluster** — Create a local Kubernetes cluster with Kind
2. **Install Flux** — Deploy Flux (source-controller, kustomize-controller, helm-controller)
3. **Deploy application** — Apply the app's `helmrelease/` kustomization with `releaseName` and `releaseNamespace` substitution
4. **Verify HelmRelease** — Assert the HelmRelease reaches `Ready` (and for upgrade tests, `UpgradeSucceeded`)

### Test Types

| Test | Description |
|------|-------------|
| **Install** | Deploy app from `applications/<app>/<version>/helmrelease`, verify HelmRelease Ready |
| **Upgrade** | Install previous version → verify Ready → upgrade to latest → verify UpgradeSucceeded |

## Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Docker](https://docs.docker.com/get-docker/) (for Kind)
- [just](https://just.systems/) (optional, for `just apptests`)

## Setup

Run from the repo root:

```bash
./catalog-workflow.sh setup
```

This runs `go mod tidy` to ensure dependencies are available. Tests are specific to this catalog (applications/ layout). No external clone.

## Running Tests

```bash
# Via justfile (from repo root)
just apptests              # run all tests
just apptests-app podinfo  # run tests for one app
just apptests-install      # install-label tests only
just apptests-upgrade      # upgrade-label tests only
just apptests-tidy         # go mod tidy
```

Or directly:

```bash
cd apptests
go test ./suites/ -v -run podinfo
go test ./suites/ -v -ginkgo.label-filter="install && podinfo"
```

## Structure

```
apptests/
├── appscenarios/       # AppScenario implementations (Install, InstallPreviousVersion, Upgrade)
│   ├── constant/       # Shared constants (DEFAULT_NAMESPACE, POLL_INTERVAL)
│   └── podinfo.go      # Podinfo scenario
├── suites/             # Ginkgo test specs
│   ├── suites_test.go  # Kind + Flux setup
│   └── podinfo_test.go # Podinfo install/upgrade tests
├── utils/              # Helpers (AbsolutePathTo, GetPrevVAppsUpgradePath)
├── go.mod
└── main.go
```

## Adding Tests for New Apps

1. Implement `AppScenario` in `appscenarios/<app>.go` (see [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md))
2. Add `suites/<app>_test.go` with Install and Upgrade specs
3. Or run `./catalog-workflow.sh add-tests --appname <app>` for placeholder scaffolding

## Reference

- [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md) — Full guide: AppScenario interface, dependencies, options
