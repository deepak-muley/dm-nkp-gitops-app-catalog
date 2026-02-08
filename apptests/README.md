# App Tests

Ginkgo/Kind integration tests for NKP catalog applications. Tests spin up a Kind cluster, install Flux, deploy each app via its HelmRelease, and verify the deployment succeeds.

## Two ways to run apptests

- **Per-app suite** (`suites/`, `appscenarios/`) — One test file and one scenario per app (e.g. `podinfo_test.go` + `appscenarios/podinfo.go`). Add a new app by implementing `AppScenario` and a `Describe` block.
- **Catalog apptests** (`catalog-apptests/` at repo root) — A single template runs **install** (and **upgrade** when an app has ≥2 versions) for **every** app under `applications/`. No per-app code: discovery scans `applications/<name>/<version>/` and the common scenario applies the same install/upgrade flow. See [catalog-apptests/README.md](../catalog-apptests/README.md).

Both suites can be run side by side. Existing apptests are unchanged.

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
| **Upgrade** | Install previous version → verify Ready → upgrade to latest → verify UpgradeSucceeded. **Requires at least two versions** of the app in `applications/<app>/` (e.g. podinfo has 6.9.3 and 6.9.4). |
| **Multicluster** | Uses [kommander-applications environment](https://github.com/mesosphere/kommander-applications/blob/main/apptests/environment/environment.go) `ProvisionMultiCluster` + `InstallLatestFluxOnWorkload`; deploy app to workload cluster and verify there. Sample: **podinfo** (`suites/podinfo_multicluster_test.go`, `appscenarios/podinfo_multicluster.go`). Run with label `multicluster` or `podinfo && multicluster`. |

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

### Choosing what to run: install vs upgrade vs both (labels)

Tests are tagged with Ginkgo labels: **`podinfo`** (app name), **`install`**, **`upgrade`**. You can run all specs, one app, or filter by label (install only, upgrade only, or both for one app).

| Goal | catalog-workflow.sh | just | go test (from apptests/) |
|------|----------------------|------|---------------------------|
| **Per-app suite: all** (install + upgrade) | `./catalog-workflow.sh test` | `just apptests` | `go test ./suites/ -v -timeout 45m` |
| **Per-app: one app** | `./catalog-workflow.sh test --appname podinfo` | `just apptests-app podinfo` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="podinfo"` |
| **Catalog apptests: all apps** | `./catalog-workflow.sh test --templated` | `just apptests-templated` | `cd catalog-apptests && go test . -v -timeout 45m` |
| **Catalog apptests: one app** | `./catalog-workflow.sh test --templated --appname podinfo` | `just apptests-templated-app podinfo` | `cd catalog-apptests && go test . -v -timeout 45m -ginkgo.label-filter="appname=podinfo"` |
| **Catalog apptests: multicluster** (install each app on workload cluster) | `./catalog-workflow.sh test --templated --label multicluster` | `just apptests-templated-label multicluster` | `cd catalog-apptests && go test . -v -timeout 45m -ginkgo.label-filter="multicluster"` |
| **Both suites** | `./catalog-workflow.sh test --all-suites` | (run `apptests` then `apptests-templated`) | — |
| **Install only** (per-app) | `./catalog-workflow.sh test --label install` | `just apptests-install` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="install"` |
| **Upgrade only** (per-app) | `./catalog-workflow.sh test --label upgrade` | `just apptests-upgrade` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="upgrade"` |
| **One app, install only** | `./catalog-workflow.sh test --appname podinfo --label install` | `just apptests-label "podinfo && install"` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="podinfo && install"` |
| **Custom label** (if you add one) | `./catalog-workflow.sh test --label mylabel` | `just apptests-label "mylabel"` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="mylabel"` |

Ginkgo label expressions: use `&&` (and), `\|\|` (or), `!` (not), e.g. `"install || upgrade"` or `"podinfo && !upgrade"`.

```bash
# Via catalog-workflow (from repo root)
./catalog-workflow.sh test
./catalog-workflow.sh test --appname podinfo
./catalog-workflow.sh test --templated              # catalog-apptests (all apps)
./catalog-workflow.sh test --templated --appname podinfo
./catalog-workflow.sh test --all-suites            # both per-app and catalog-apptests
./catalog-workflow.sh test --label install
./catalog-workflow.sh test --label upgrade

# Via justfile (from repo root)
just apptests
just apptests-app podinfo
just apptests-templated
just apptests-templated-app podinfo
just apptests-install
just apptests-upgrade
just apptests-label "podinfo && install"
just apptests-tidy
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

The **catalog-apptests** suite lives in `catalog-apptests/` at repo root (separate Go module); see [catalog-apptests/README.md](../catalog-apptests/README.md).

## Adding Tests for New Apps

1. Implement `AppScenario` in `appscenarios/<app>.go` (see [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md))
2. Add `suites/<app>_test.go` with Install and Upgrade specs
3. Or run `./catalog-workflow.sh add-tests --appname <app>` for placeholder scaffolding

## Kind cluster: delete vs keep

- **Default: cluster is deleted** after each test run. The suite’s `AfterEach` calls `env.Destroy(ctx)` so the Kind cluster (e.g. `kommanderapptest`) is torn down and doesn’t linger.
- **Keep the cluster for debugging:** set `SKIP_CLUSTER_TEARDOWN=1` when running tests. The cluster will remain so you can inspect it with `kubectl` or `kind get clusters`. Remember to delete it when done: `kind delete cluster --name kommanderapptest`.
- **Cluster name** — Defaults to `kommanderapptest`; to use a custom name see [Custom Kind cluster name](#custom-kind-cluster-name) below.

## Custom Kind cluster name

The framework's [kind package](https://github.com/mesosphere/kommander-applications/blob/main/apptests/kind/kind.go) accepts a cluster name in `CreateCluster(ctx, name)` but uses `"kommanderapptest"` when `name` is empty. The environment package always calls it with `""`, so the name is not configurable without a small override.

To use a **custom cluster name**:

1. Set **`KIND_CLUSTER_NAME`** when running tests:
   ```bash
   export KIND_CLUSTER_NAME=my-nkp-test
   ./catalog-workflow.sh test --templated
   ```
2. In **`apptests/go.mod`**, uncomment the replace line at the bottom so the `kind` package is taken from **`apptests/kindoverride`** (this repo includes it):
   ```go
   replace github.com/mesosphere/kommander-applications/apptests/kind => ./kindoverride
   ```
3. Run `cd apptests && go mod tidy`, then run the tests again. The override uses `os.Getenv("KIND_CLUSTER_NAME")` when the framework passes an empty name.

Without the replace, `KIND_CLUSTER_NAME` has no effect.

## Troubleshooting

- **Test timed out after 10m0s** — The default `go test` timeout is 10 minutes. We use `-timeout 45m` in the justfile and catalog-workflow so Kind + Flux + MetalLB (from the test framework) have time to come up. If you run `go test` by hand, add `-timeout 45m`.
- **MetalLB / metallb-speaker not ready** — The test environment (kommander-applications/apptests) provisions a Kind cluster and waits for MetalLB. If the DaemonSet never becomes ready, ensure Docker has enough resources (CPU/memory) and that nothing is blocking the speaker. The longer timeout gives the cluster more time to stabilize.
- **Release installed (metallb-0.13.7) but speaker does nothing** — Helm says the release is installed, but the `metallb-speaker` DaemonSet stays at 0/1 ready. The blocker is usually the **pods**, not the install. Run `./apptests/debug-metallb.sh` (or the kubectl commands below) and check:
  - **Pod status:** `kubectl describe pod -n metallb-system <speaker-pod>` — look at **State** and **Events** (e.g. `FailedScheduling`, `CreateContainerConfigError`, `CrashLoopBackOff`).
  - **Pod security:** MetalLB speaker often needs `hostNetwork`, `hostPorts`, or `NET_RAW`. If the namespace has Pod Security Standards (restricted/baseline), the pod can be **rejected** or **admitted but not scheduled**. Events may show "forbidden" or capability errors.
  - **Secrets:** Speaker needs the `metallb-memberlist` secret; if it’s missing, pods fail with mount/ConfigMap errors.
  - **Logs:** `kubectl logs -n metallb-system -l component=speaker --tail=100` — look for bind/probe or permission errors.
  On Kind (especially macOS/Docker Desktop), the speaker sometimes never becomes ready due to these constraints; the test framework installs MetalLB and we don’t control that. Workarounds are limited to fixing the framework’s cluster config or accepting that apptests may fail in some environments.
  - **Debug:** Keep the cluster and inspect MetalLB:
    1. Run tests with cluster left up: `SKIP_CLUSTER_TEARDOWN=1 ./catalog-workflow.sh test --appname podinfo` (or let it time out; the cluster may still exist).
    2. From repo root: `./apptests/debug-metallb.sh` — prints nodes, DaemonSet, pods, describe, logs, and events for `metallb-system/metallb-speaker`.
    3. Or use Kind + kubectl by hand:
      ```bash
      export KUBECONFIG="$(kind get kubeconfig --name kommanderapptest)"
      kubectl get nodes
      kubectl get pods -n metallb-system -o wide
      kubectl describe daemonset -n metallb-system metallb-speaker
      kubectl describe pod -n metallb-system -l app=metallb,component=speaker
      kubectl logs -n metallb-system -l app=metallb,component=speaker --tail=100
      kubectl get events -n metallb-system --sort-by='.lastTimestamp'
      ```
    4. When done, delete the cluster: `kind delete cluster --name kommanderapptest`.

## Reference

- [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md) — Full guide: AppScenario interface, dependencies, options
