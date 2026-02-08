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
| **Upgrade** | Install previous version → verify Ready → upgrade to latest → verify UpgradeSucceeded. **Requires at least two versions** of the app in `applications/<app>/` (e.g. podinfo has 6.9.3 and 6.9.4). |

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
| **All tests** (install + upgrade, all apps) | `./catalog-workflow.sh test` | `just apptests` | `go test ./suites/ -v -timeout 45m` |
| **One app** (install + upgrade) | `./catalog-workflow.sh test --appname podinfo` | `just apptests-app podinfo` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="podinfo"` |
| **Install only** (all apps) | `./catalog-workflow.sh test --label install` | `just apptests-install` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="install"` |
| **Upgrade only** (all apps) | `./catalog-workflow.sh test --label upgrade` | `just apptests-upgrade` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="upgrade"` |
| **One app, install only** | `./catalog-workflow.sh test --appname podinfo --label install` | `just apptests-label "podinfo && install"` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="podinfo && install"` |
| **One app, upgrade only** | `./catalog-workflow.sh test --appname podinfo --label upgrade` | `just apptests-label "podinfo && upgrade"` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="podinfo && upgrade"` |
| **Custom label** (if you add one) | `./catalog-workflow.sh test --label mylabel` | `just apptests-label "mylabel"` | `go test ./suites/ -v -timeout 45m -ginkgo.label-filter="mylabel"` |

Ginkgo label expressions: use `&&` (and), `\|\|` (or), `!` (not), e.g. `"install || upgrade"` or `"podinfo && !upgrade"`.

```bash
# Via catalog-workflow (from repo root)
./catalog-workflow.sh test
./catalog-workflow.sh test --appname podinfo
./catalog-workflow.sh test --label install
./catalog-workflow.sh test --label upgrade
./catalog-workflow.sh test --appname podinfo --label install

# Via justfile (from repo root)
just apptests
just apptests-app podinfo
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

## Adding Tests for New Apps

1. Implement `AppScenario` in `appscenarios/<app>.go` (see [docs/APP-TESTS-GUIDE.md](../docs/APP-TESTS-GUIDE.md))
2. Add `suites/<app>_test.go` with Install and Upgrade specs
3. Or run `./catalog-workflow.sh add-tests --appname <app>` for placeholder scaffolding

## Kind cluster: delete vs keep

- **Default: cluster is deleted** after each test run. The suite’s `AfterEach` calls `env.Destroy(ctx)` so the Kind cluster (e.g. `kommanderapptest`) is torn down and doesn’t linger.
- **Keep the cluster for debugging:** set `SKIP_CLUSTER_TEARDOWN=1` when running tests. The cluster will remain so you can inspect it with `kubectl` or `kind get clusters`. Remember to delete it when done: `kind delete cluster --name kommanderapptest`.
- **Cluster name** — The name (e.g. `kommanderapptest`) is set by the test framework (kommander-applications/apptests), not by this repo. To use a different name you’d need to check whether the framework’s `environment.Env` or Provision API supports it.

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
