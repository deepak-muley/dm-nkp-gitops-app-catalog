# Kubeflow in NKP Catalog

This document describes how to add Kubeflow components to this catalog, constraints, and **exact workflow commands** for each component.

## NKP and Flux Constraints

- **NKP does not support HelmRepository** in catalog bundles; only OCI Helm charts via `HelmRelease` are supported.
- For Helm charts: `add-application.sh` requires an **OCI path** (`--ocirepo oci://...`) or `--helmrepo` + `--ocipush`.
- For **Kustomize-based** components: use `--kustomize --gitrepo --path` to add GitRepository + Flux Kustomization.

---

## Exact Workflow Commands by Component

Use `./catalog-workflow.sh add-app` (or `./catalog-workflow.sh add-app`) with these exact arguments. All Kustomize paths reference [kubeflow/manifests](https://github.com/kubeflow/manifests). Use `--ref master` for latest or `--ref release-v1.10` for a stable release.

### Helm (OCI) — KServe

| Component | Version | Command |
|-----------|---------|---------|
| **KServe** | 0.16.0 | `./catalog-workflow.sh add-app --appname kserve --version 0.16.0 --ocirepo oci://ghcr.io/kserve/charts/kserve` |
| **KServe CRD** | 0.16.0 | `./catalog-workflow.sh add-app --appname kserve-crd --version 0.16.0 --ocirepo oci://ghcr.io/kserve/charts/kserve-crd` |

**Note:** GHCR publishes these charts with a `v` prefix (e.g. `v0.16.0`). After add-app, set `spec.ref.tag` to `v0.16.0` in each app's `helmrelease/helmrelease.yaml` or validation will fail with MANIFEST_UNKNOWN.

### Kustomize — [kubeflow/manifests](https://github.com/kubeflow/manifests)

| Component | Version | Path | Command |
|-----------|---------|------|---------|
| **Katib** | 0.19.0 | `applications/katib/upstream/installs/katib-with-kubeflow` | `./catalog-workflow.sh add-app --appname katib --version 0.19.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/katib/upstream/installs/katib-with-kubeflow --ref master` |
| **Kubeflow Pipelines** (DB mode) | 2.15.0 | `applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user` | `./catalog-workflow.sh add-app --appname kubeflow-pipelines --version 2.15.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user --ref master` |
| **Kubeflow Pipelines** (K8s native) | 2.15.0 | `applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user-k8s-native` | `./catalog-workflow.sh add-app --appname kubeflow-pipelines-k8s-native --version 2.15.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user-k8s-native --ref master` |
| **Model Registry** | 0.3.6 | Standalone repo: `manifests/kustomize/overlays/db` | `./catalog-workflow.sh add-app --appname kubeflow-model-registry --version 0.3.6 --kustomize --gitrepo https://github.com/kubeflow/model-registry --path ./manifests/kustomize/overlays/db --ref v0.3.6` — Uses [kubeflow/model-registry](https://github.com/kubeflow/model-registry) (not manifests) so validation does not fail on missing `options/catalog/options/istio`. |
| **Spark Operator** (Kustomize) | 2.4.0 | `applications/spark/spark-operator/overlays/kubeflow` | `./catalog-workflow.sh add-app --appname spark-operator --version 2.4.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/spark/spark-operator/overlays/kubeflow --ref master` |
| **Trainer** (training-operator v2) | 2.1.0 | `applications/trainer/upstream/overlays/kubeflow-platform` | `./catalog-workflow.sh add-app --appname trainer --version 2.1.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/trainer/upstream/overlays/kubeflow-platform --ref master` |
| **Training Operator** (v1) | 1.9.2 | `applications/training-operator/upstream/overlays/kubeflow` | `./catalog-workflow.sh add-app --appname training-operator --version 1.9.2 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/training-operator/upstream/overlays/kubeflow --ref master` |
| **Central Dashboard** | 1.10.0 | `applications/centraldashboard/overlays/oauth2-proxy` | `./catalog-workflow.sh add-app --appname kubeflow-central-dashboard --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/centraldashboard/overlays/oauth2-proxy --ref master` |
| **KServe** (Kustomize) | 0.15.2 | `applications/kserve/kserve` | `./catalog-workflow.sh add-app --appname kserve-kustomize --version 0.15.2 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/kserve/kserve --ref master` |
| **Jupyter Notebook Controller** | 1.10.0 | `applications/jupyter/notebook-controller/upstream/overlays/kubeflow` | `./catalog-workflow.sh add-app --appname jupyter-notebook-controller --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/jupyter/notebook-controller/upstream/overlays/kubeflow --ref master` |
| **Jupyter Web App** | 1.10.0 | `applications/jupyter/jupyter-web-app/upstream/overlays/istio` | `./catalog-workflow.sh add-app --appname jupyter-web-app --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/jupyter/jupyter-web-app/upstream/overlays/istio --ref master` |
| **Profiles + KFAM** | 1.10.0 | `applications/profiles/upstream/overlays/kubeflow` | `./catalog-workflow.sh add-app --appname kubeflow-profiles --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/profiles/upstream/overlays/kubeflow --ref master` |
| **Admission Webhook** | 1.10.0 | `applications/admission-webhook/upstream/overlays/cert-manager` | `./catalog-workflow.sh add-app --appname kubeflow-admission-webhook --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/admission-webhook/upstream/overlays/cert-manager --ref master` |
| **PVC Viewer Controller** | 1.10.0 | `applications/pvcviewer-controller/upstream/base` | `./catalog-workflow.sh add-app --appname pvcviewer-controller --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/pvcviewer-controller/upstream/base --ref master` |
| **Volumes Web App** | 1.10.0 | `applications/volumes-web-app/upstream/overlays/istio` | `./catalog-workflow.sh add-app --appname volumes-web-app --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/volumes-web-app/upstream/overlays/istio --ref master` |
| **Tensorboard Controller** | 1.10.0 | `applications/tensorboard/tensorboard-controller/upstream/overlays/kubeflow` | `./catalog-workflow.sh add-app --appname tensorboard-controller --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/tensorboard/tensorboard-controller/upstream/overlays/kubeflow --ref master` |
| **Tensorboards Web App** | 1.10.0 | `applications/tensorboard/tensorboards-web-app/upstream/overlays/istio` | `./catalog-workflow.sh add-app --appname tensorboards-web-app --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/tensorboard/tensorboards-web-app/upstream/overlays/istio --ref master` |

### Helm Repo → OCI — Spark Operator (alternative)

Spark Operator is also available via Helm. Use `--helmrepo` and `--ocipush`:

```bash
./catalog-workflow.sh add-app --appname spark-operator --version 1.1.31 \
  --helmrepo spark-operator/spark-operator \
  --ocipush oci://ghcr.io/<your-org>/spark-operator \
  --helmrepo-url https://kubeflow.github.io/spark-operator
```

---

## Option Flags for Kustomize

- **`--gitrepo`** — Git repo URL: `https://github.com/kubeflow/manifests`
- **`--path`** — Path within repo to the kustomization.yaml directory (see table above)
- **`--ref`** — Branch or tag: `master` (latest) or `release-v1.10` (stable)

Use `--force` to overwrite existing application directories.

---

## Dependencies and Order

Many Kubeflow components depend on others (Istio, cert-manager, Knative, Profiles, etc.). Install in this order when building a full stack:

1. **common** (Istio, cert-manager, Knative, Dex, OAuth2-proxy, etc.) — typically via Nutanix kubeflow-manifests or manual kustomize
2. **Profiles + KFAM**
3. **Central Dashboard**, **Admission Webhook**, **PVC Viewer**
4. **Pipelines**, **Katib**, **KServe**, **Trainer**, **Spark Operator**, **Model Registry**
5. **Jupyter**, **Tensorboard**, **Volumes Web App**

For standalone components (e.g., Katib alone), ensure required CRDs and dependencies exist.

---

## Full Kubeflow Platform (Outside Catalog)

The full Kubeflow AI Reference Platform (Nutanix kubeflow-manifests) uses Kustomize and is **not** deployable via this catalog's add-app flow.

Install it separately using [Nutanix Kubeflow Manifests](https://nutanix.github.io/kubeflow-manifests/docs/install-kubeflow/):

```bash
git clone -b release-v1.10.1 https://github.com/nutanix/kubeflow-manifests.git
cd kubeflow-manifests
make install-nkp-kubeflow
```

---

## Post-Add Steps

After adding any component:

```bash
./catalog-workflow.sh validate
# Update metadata.yaml (description, supportLink)
./catalog-workflow.sh build-push --tag v0.1.0
```
