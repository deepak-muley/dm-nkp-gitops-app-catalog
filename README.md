# NKP App Catalog

A custom app catalog for [Nutanix Kubernetes Platform (NKP)](https://www.nutanix.com/products/kubernetes-platform). Applications are Flux/Kustomize manifests with OCI Helm charts and catalog metadata for the NKP UI.

## Prerequisites

- [nkp](https://github.com/nutanix-cloud-native/nkp) CLI (or `./nkp` in repo)
- [helm](https://helm.sh/docs/intro/install/) 3.8+
- OCI registry login (e.g. `docker login ghcr.io`)

## Quick Start

```bash
# Add an app (OCI chart)
./catalog-workflow.sh add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo

# Add an app (Helm repo → OCI)
./catalog-workflow.sh add-app --appname kyverno --version 3.6.1 --helmrepo kyverno/kyverno --ocipush oci://ghcr.io/YOUR_ORG/kyverno --helmrepo-url https://kyverno.github.io/kyverno/

# Validate catalog
./catalog-workflow.sh validate

# Build and push bundle
./catalog-workflow.sh build-push --tag v0.1.0
```

## Command Reference

| Command | Description |
|---------|-------------|
| `./catalog-workflow.sh add-app --appname <name> --version <ver> --ocirepo oci://...` | Add app from OCI chart |
| `./catalog-workflow.sh add-app --appname <name> --version <ver> --helmrepo <repo/chart> --ocipush oci://...` | Add app (pull from Helm repo, push to OCI) |
| `./catalog-workflow.sh validate` | Validate catalog (nkp validate + OCI login) |
| `./catalog-workflow.sh check-versions [--appname <name>\|--all]` | Check latest chart versions; recommend add-app commands |
| `./catalog-workflow.sh build-push --tag <version>` | Build and push catalog bundle |
| `./catalog-workflow.sh add-tests --appname <name>` | Create apptest placeholders |
| `./catalog-workflow.sh test [--appname <app>] [--label install/upgrade]` | Run apptests |
| `./catalog-workflow.sh setup` | Run setup (tools + apptests go mod tidy) |
| `./catalog-workflow.sh validate` | Validate only |
| `./catalog-workflow.sh build-push --tag <tag>` | Build and push (loads credentials from `.env.local`) |

## Applications

| Category | Apps |
|----------|------|
| **Infrastructure** | cert-manager, traefik, letsencrypt-clusterissuer |
| **Observability** | kube-prometheus-stack, loki, tempo, opentelemetry-operator |
| **Security** | kyverno, kubescape-operator, vault, oauth2-proxy |
| **Platform** | karmada-operator, kro, agentgateway |

Full list with add commands: [docs/ADD-APPLICATION-COMMANDS.md](docs/ADD-APPLICATION-COMMANDS.md)

## Documentation

| Doc | Description |
|-----|-------------|
| [docs/CATALOG-APPS-GUIDE.md](docs/CATALOG-APPS-GUIDE.md) | **Learn each app** — what it is, what it does, UI/dashboard, how to use it |
| [docs/ADD-APPLICATION-COMMANDS.md](docs/ADD-APPLICATION-COMMANDS.md) | Commands for each application |
| [docs/CATALOG-SOURCE.md](docs/CATALOG-SOURCE.md) | Per-app source metadata for Helm-repo → OCI apps (check-versions) |
| [docs/CATALOG-WORKFLOW.md](docs/CATALOG-WORKFLOW.md) | catalog-workflow.sh usage |
| [docs/REFERENCE.md](docs/REFERENCE.md) | NKP workflow, bundle push, helm values |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Full stack deployment order |
| [docs/KUBEFLOW-CATALOG.md](docs/KUBEFLOW-CATALOG.md) | Adding Kubeflow components |
| [docs/APP-TESTS-GUIDE.md](docs/APP-TESTS-GUIDE.md) | Ginkgo/Kind integration tests |
| [docs/TEST-CI-LOCALLY.md](docs/TEST-CI-LOCALLY.md) | How to run CI steps locally |

## Structure

```
applications/<app>/<version>/
├── metadata.yaml       # catalog.nkp.nutanix.com/v1/application-metadata
├── helmrelease.yaml
├── kustomization.yaml
└── helmrelease/        # Flux OCIRepository + HelmRelease
```

## Apptests (Integration Tests)

```bash
./catalog-workflow.sh setup  # one-time setup (go mod tidy)
./catalog-workflow.sh test   # run all tests (requires Docker)
./catalog-workflow.sh test --appname podinfo
```

Tests are specific to this catalog. **CI runs apptests on every push** (after validate). To run the same locally: `./catalog-workflow.sh ci-local`. See [docs/TEST-CI-LOCALLY.md](docs/TEST-CI-LOCALLY.md), [apptests/README.md](apptests/README.md), and [docs/APP-TESTS-GUIDE.md](docs/APP-TESTS-GUIDE.md).

## Pre-push validation

A Git pre-push hook runs `./catalog-workflow.sh validate` before each push so broken catalog state is not pushed. Install it once:

```bash
cp .githooks/pre-push .git/hooks/pre-push && chmod +x .git/hooks/pre-push
```

The hook also blocks the push if any sensitive files (e.g. `.env.local`, `setup-credentials.sh`) are tracked.

## Contributing

1. Add apps via `./catalog-workflow.sh add-app`
2. Run `./catalog-workflow.sh validate` before committing (or rely on the pre-push hook)
3. See [docs/ADD-APPLICATION-COMMANDS.md](docs/ADD-APPLICATION-COMMANDS.md) for conventions
