# NKP App Catalog — Reference

Detailed reference for building, pushing, and importing NKP catalog bundles. For quick commands, see [README](../README.md).

To **learn what each app is, what it does, whether it has a UI, and how to use it**, see [CATALOG-APPS-GUIDE.md](CATALOG-APPS-GUIDE.md).

## Prerequisites

- [nkp](https://github.com/nutanix-cloud-native/nkp) CLI (or `./nkp` in repo)
- [helm](https://helm.sh/docs/intro/install/) 3.8+
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- Docker/ORAS login to your OCI registry (e.g. `docker login ghcr.io`)

## Repository Layout

```
applications/<app>/<version>/
├── metadata.yaml          # catalog.nkp.nutanix.com/v1/application-metadata — see [APPLICATION-METADATA-FIELDS.md](APPLICATION-METADATA-FIELDS.md)
├── helmrelease.yaml
├── kustomization.yaml
└── helmrelease/           # Flux OCIRepository + HelmRelease
```

## NKP Workflow (Manual Steps)

### 1. Generate App Structure

```bash
nkp generate catalog-repository --apps=podinfo=6.9.4
```

### 2. Validate

```bash
nkp validate catalog-repository --repo-dir=.
```

### 3. Create Bundle

```bash
nkp create catalog-bundle --collection-tag v0.1.0
```

### 4. Push Bundle to OCI

```bash
nkp push bundle --bundle ./dm-nkp-gitops-app-catalog.tar \
  --to-registry oci://ghcr.io/<org>/nkp-custom-apps-catalog \
  --to-registry-username <user> --to-registry-password <token>
```

### 5. Import in NKP

```bash
nkp create catalog-collection --url oci://ghcr.io/<org>/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection \
  --tag v0.1.0 --workspace <workspace>
```

## Automated Build and Push

Use `./catalog-workflow.sh build-push` to automate validation, bundle creation, and pushing:

```bash
# Setup credentials
cp setup-credentials.sh.example setup-credentials.sh
# Edit setup-credentials.sh with GHCR_USERNAME, GHCR_PASSWORD
source setup-credentials.sh

# Build and push
./build-and-push.sh v0.1.0
```

## OCI Registry Login

For private OCI registries (e.g. ghcr.io):

```bash
echo "YOUR_GITHUB_PAT" | docker login ghcr.io -u "YOUR_USERNAME" --password-stdin
# or
docker login ghcr.io
```

GitHub PAT scopes: `write:packages`, `read:packages`.

## Helm Chart → OCI (Charts Not Yet in OCI)

NKP requires OCI Helm charts. For charts on traditional Helm repos, use `--helmrepo` and `--ocipush`:

```bash
./catalog-workflow.sh add-app --appname kyverno --version 3.6.1 \
  --helmrepo kyverno/kyverno \
  --ocipush oci://ghcr.io/<org>/kyverno \
  --helmrepo-url https://kyverno.github.io/kyverno/
```

See [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md) for more examples.

## Helm Values

```bash
# OCI chart
helm show values oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack

# Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm show values prometheus-community/kube-prometheus-stack

# Deployed release
helm get values <release-name> -n <namespace>
```

## Nutanix NKP Documentation

- [Custom Apps](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-custom-apps-c.html)
- [Partner Catalog](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-partner-catalog-in-nkp-c.html)
- [Workspace App Metadata](https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-workspace-app-metadata-c.html)
