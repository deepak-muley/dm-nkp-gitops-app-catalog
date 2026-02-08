# Catalog apptests (common template over all catalog apps)

This package runs **table-driven** tests: one common template (install, upgrade, multicluster) is applied to **every app** under `applications/`, with no per-app test code. It lives at repo root as its own Go module (`catalog-apptests/`) and **does not depend on any upstream apptests library**.

## How it works

1. **Discovery** (`discovery.go`)  
   Scans `applications/<name>/` and collects all apps that have at least one version directory. Versions are sorted so "previous" and "latest" are well-defined.

2. **Cluster API** (`cluster.go`)  
   - **KindCluster.Create(ctx, network, name)** – creates the first cluster (mgmt) on the given Docker network.
   - **KindCluster.CreateFromParent(ctx, parent, name)** – returns the workload cluster (parent must have been created with `Create`).
   - **KindCluster.CreateSingle(ctx, network, name)** – creates one cluster only (no workload).
   - **cluster.Install(FluxApp)** – installs Flux (source-, kustomize-, helm-controller).
   - **cluster.Install(catalogApp)** – applies the app's helmrelease kustomization.
   - **cluster.Destroy()** – tears down the cluster(s).

3. **Framework** (`framework/`)  
   Self-contained helpers: Docker network, Kind cluster create/delete, K8s client from kubeconfig, Flux install (flux2 manifestgen + ssa apply), kustomize build + envsubst + apply. No dependency on `github.com/mesosphere/kommander-applications/apptests`.

4. **Scenario** (`scenario.go`)  
   **CatalogApp** implements install/upgrade against a **Cluster** (Install, InstallPreviousVersion, Upgrade).

5. **Suite** (`suite_test.go`)  
   Single-cluster: for each app, install latest (and upgrade when ≥2 versions). Multicluster: mgmt + workload1 + workload2, install Flux and catalog app on each.

## Example (desired API)

```go
network, _ := framework.EnsureDockerNetworkExist(ctx, "", false)

mgmt, _ := KindCluster.Create(ctx, network, "mgmt")
workload1, _ := KindCluster.CreateFromParent(ctx, mgmt, "workload1")
workload2, _ := KindCluster.CreateFromParent(ctx, mgmt, "workload2")

mgmt.Install(FluxApp)
workload1.Install(FluxApp)
workload2.Install(FluxApp)

catalogApp := &CatalogApp{AppName: "podinfo", VersionToInstall: ""}
mgmt.Install(catalogApp)
workload1.Install(catalogApp)
workload2.Install(catalogApp)

defer mgmt.Destroy()
```

## Running

From **repo root**:

```bash
./catalog-workflow.sh test --catalog-apptests
./catalog-workflow.sh test --catalog-apptests --appname podinfo
./catalog-workflow.sh test --catalog-apptests -- -ginkgo.label-filter="multicluster"
```

With **go test** (from repo root):

```bash
cd catalog-apptests
go mod tidy
go test . -v -timeout 45m
go test . -v -timeout 45m -ginkgo.label-filter="appname=podinfo"
```

## Layout

```
catalog-apptests/
├── go.mod
├── framework/           # Self-contained: network, kind, client, flux, kustomize, scheme
│   ├── network.go
│   ├── kind.go
│   ├── client.go
│   ├── scheme.go
│   ├── flux.go
│   └── kustomize.go
├── cluster.go          # Cluster interface; KindCluster.Create / CreateSingle / CreateFromParent
├── scenario.go         # CatalogApp
├── discovery.go
├── constants.go
├── suite_test.go
└── README.md
```

## Dependencies

Direct dependencies (no upstream apptests): Docker client, Kind, Flux2 (manifestgen + install), fluxcd/pkg/ssa, fluxcd/source-controller and kustomize-controller APIs, kustomize, envsubst, ginkgo/gomega, controller-runtime.
