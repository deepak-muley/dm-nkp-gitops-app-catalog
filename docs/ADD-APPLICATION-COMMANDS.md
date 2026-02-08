# add-application.sh Commands Reference

Reference for recreating this catalog from scratch using `./catalog-workflow.sh add-app`. Data sourced from existing applications.

## Modes: Helm vs Kustomize

- **Helm (OCI)** — `--ocirepo` — Use when the chart is already published to an OCI registry.
- **Helm (repo)** — `--helmrepo` + `--ocipush` — Use when the chart is only on a Helm repo: the script pulls, pushes to OCI, then adds the app.
- **Kustomize** — `--kustomize` + `--gitrepo` + `--path` — Use for Kubeflow-style Kustomize-based deployments (GitRepository + Flux Kustomization).

## Usage

```bash
# Direct OCI (chart already in OCI)
./catalog-workflow.sh add-app --appname <name> --version <version> --ocirepo oci://<registry>/path/chart [--force]

# Helm repo → OCI (script pulls and pushes)
./catalog-workflow.sh add-app --appname <name> --version <version> --helmrepo <repo>/<chart> --ocipush oci://<registry>/path [--helmrepo-url <url>] [--force]

# Kustomize (Kubeflow-style: GitRepository + Flux Kustomization)
./catalog-workflow.sh add-app --appname <name> --version <version> --kustomize --gitrepo <url> --path <path> [--ref <branch|tag>] [--force]

# Via workflow script
./catalog-workflow.sh add-app --appname <name> --version <version> --ocirepo oci://...
./catalog-workflow.sh add-app --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default --ref release-v1.10
```

Use `--force` to overwrite existing application directories without confirmation.

**Workflow script:** `./catalog-workflow.sh` orchestrates add-app, validate, add-tests, build-push. Use `./catalog-workflow.sh --help` for all options.

---

## Applications Using add-application.sh (OCI Helm Charts)

### Infrastructure

| App | Version | Command |
|-----|---------|---------|
| cert-manager | 1.19.2 | `./catalog-workflow.sh add-app --appname cert-manager --version 1.19.2 --ocirepo oci://quay.io/jetstack/charts/cert-manager` |
| traefik | 38.0.2 | `./catalog-workflow.sh add-app --appname traefik --version 38.0.2 --ocirepo oci://ghcr.io/traefik/helm/traefik` |

### Observability

| App | Version | Command |
|-----|---------|---------|
| kube-prometheus-stack | 80.13.3 | `./catalog-workflow.sh add-app --appname kube-prometheus-stack --version 80.13.3 --ocirepo oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack` |
| loki | 6.46.0 | `./catalog-workflow.sh add-app --appname loki --version 6.46.0 --ocirepo oci://ghcr.io/grafana/helm-charts/loki` |
| tempo | 1.21.1 | `./catalog-workflow.sh add-app --appname tempo --version 1.21.1 --ocirepo oci://ghcr.io/grafana/helm-charts/tempo` |
| opentelemetry-collector | 0.140.1 | `./catalog-workflow.sh add-app --appname opentelemetry-collector --version 0.140.1 --ocirepo oci://ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-collector` |
| opentelemetry-operator | 0.140.0 | `./catalog-workflow.sh add-app --appname opentelemetry-operator --version 0.140.0 --ocirepo oci://ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-operator` |
| clickstack | 1.1.1 | `./catalog-workflow.sh add-app --appname clickstack --version 1.1.1 --helmrepo clickstack/clickstack --helmrepo-url https://clickhouse.github.io/ClickStack-helm-charts --ocipush oci://ghcr.io/deepak-muley/charts/clickstack` |

### Demo / General

| App | Version | Command |
|-----|---------|---------|
| podinfo | 6.9.4 | `./catalog-workflow.sh add-app --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo` |

### Security

| App | Version | Command |
|-----|---------|---------|
| oauth2-proxy | 8.0.2 | `./catalog-workflow.sh add-app --appname oauth2-proxy --version 8.0.2 --ocirepo oci://registry-1.docker.io/bitnamicharts/oauth2-proxy` |

### Multi-Cluster / Platform

| App | Version | Command |
|-----|---------|---------|
| kro | 0.7.1 | `./catalog-workflow.sh add-app --appname kro --version 0.7.1 --ocirepo oci://registry.k8s.io/kro/charts/kro` |

### AI / Infrastructure (from kgateway-dev)

| App | Version | Command |
|-----|---------|---------|
| agentgateway | 2.1.2 | `./catalog-workflow.sh add-app --appname agentgateway --version 2.1.2 --ocirepo oci://ghcr.io/kgateway-dev/charts/agentgateway` |
| | | *Note: Chart uses tag `v2.2.0-main`; update `ref.tag` in helmrelease if needed.* |

### AI / HPC — Slurm (Slinky, OCI)

| App | Version | Command |
|-----|---------|---------|
| slurm-operator-crds | 1.0.0 | `./catalog-workflow.sh add-app --appname slurm-operator-crds --version 1.0.0 --ocirepo oci://ghcr.io/slinkyproject/charts/slurm-operator-crds --force` |
| slurm-operator | 1.0.0 | `./catalog-workflow.sh add-app --appname slurm-operator --version 1.0.0 --ocirepo oci://ghcr.io/slinkyproject/charts/slurm-operator --force` |
| slurm | 1.0.0 | `./catalog-workflow.sh add-app --appname slurm --version 1.0.0 --ocirepo oci://ghcr.io/slinkyproject/charts/slurm --force` |

*Install order: slurm-operator-crds → slurm-operator → slurm. Requires cert-manager. See [AI-APPS-RECOMMENDATIONS.md](AI-APPS-RECOMMENDATIONS.md).*

---

## Applications Using Helm Repo (--helmrepo + --ocipush)

These charts are on Helm repos. Use `--helmrepo` and `--ocipush` so the script pulls and pushes to OCI automatically.

| App | Version | Command |
|-----|---------|---------|
| kubescape-operator | 1.29.12 | `./catalog-workflow.sh add-app --appname kubescape-operator --version 1.29.12 --helmrepo kubescape/kubescape-operator --ocipush oci://ghcr.io/deepak-muley/kubescape-operator --helmrepo-url https://kubescape.github.io/helm-charts/` |
| kyverno | 3.6.1 | `./catalog-workflow.sh add-app --appname kyverno --version 3.6.1 --helmrepo kyverno/kyverno --ocipush oci://ghcr.io/deepak-muley/kyverno --helmrepo-url https://kyverno.github.io/kyverno/` |
| vault | 0.31.0 | `./catalog-workflow.sh add-app --appname vault --version 0.31.0 --helmrepo hashicorp/vault --ocipush oci://ghcr.io/deepak-muley/vault --helmrepo-url https://helm.releases.hashicorp.com` |
| karmada-operator | 1.16.0 | `./catalog-workflow.sh add-app --appname karmada-operator --version 1.16.0 --helmrepo karmada-charts/karmada-operator --ocipush oci://ghcr.io/deepak-muley/karmada-operator --helmrepo-url https://raw.githubusercontent.com/karmada-io/karmada/master/charts` |
| vllm | 0.1.1 | `./catalog-workflow.sh add-app --appname vllm --version 0.1.1 --helmrepo vllm/vllm --ocipush oci://ghcr.io/deepak-muley/vllm --helmrepo-url https://open-source-ai-dev.github.io/vllm-helm-chart` |
| open-webui | 10.2.1 | `./catalog-workflow.sh add-app --appname open-webui --version 10.2.1 --helmrepo open-webui/open-webui --ocipush oci://ghcr.io/deepak-muley/open-webui --helmrepo-url https://helm.openwebui.com/ --force` |
| weaviate | 17.7.0 | `./catalog-workflow.sh add-app --appname weaviate --version 17.7.0 --helmrepo weaviate/weaviate --ocipush oci://ghcr.io/deepak-muley/weaviate --helmrepo-url https://weaviate.github.io/weaviate-helm/ --force` |
| local-ai | 3.4.2 | `./catalog-workflow.sh add-app --appname local-ai --version 3.4.2 --helmrepo go-skynet/local-ai --ocipush oci://ghcr.io/deepak-muley/local-ai --helmrepo-url https://go-skynet.github.io/helm-charts/ --force` |
| langsmith | 0.13.7 | `./catalog-workflow.sh add-app --appname langsmith --version 0.13.7 --helmrepo langchain/langsmith --ocipush oci://ghcr.io/deepak-muley/langsmith --helmrepo-url https://langchain-ai.github.io/helm/ --force` |
| langgraph-cloud | 0.2.1 | `./catalog-workflow.sh add-app --appname langgraph-cloud --version 0.2.1 --helmrepo langchain/langgraph-cloud --ocipush oci://ghcr.io/deepak-muley/langgraph-cloud --helmrepo-url https://langchain-ai.github.io/helm/ --force` |
| langgraph-dataplane | 0.2.15 | `./catalog-workflow.sh add-app --appname langgraph-dataplane --version 0.2.15 --helmrepo langchain/langgraph-dataplane --ocipush oci://ghcr.io/deepak-muley/langgraph-dataplane --helmrepo-url https://langchain-ai.github.io/helm/ --force` |

**Note:** If the Helm repo is already added (`helm repo add` + `helm repo update`), you can omit `--helmrepo-url`.

For these apps, **new versions appear in the upstream Helm repo**, not in your OCI registry. To have `./catalog-workflow.sh check-versions` check the Helm repo and recommend add-app commands, add a **`.catalog-source.yaml`** in each app folder — see [docs/CATALOG-SOURCE.md](CATALOG-SOURCE.md).

**Optional AI / RAG (community charts):** Anything-LLM and RAGFlow have community Helm charts (e.g. [la-cc/anything-llm-helm-chart](https://github.com/la-cc/anything-llm-helm-chart), [fzhan/ragflow-helm](https://github.com/fzhan/ragflow-helm)). Add via `--helmrepo` once you have the Helm repo URL and chart version; see [AI-APPS-RECOMMENDATIONS.md](AI-APPS-RECOMMENDATIONS.md).

---

## Applications Using Kustomize (--kustomize)

For Kubeflow-style Kustomize-based deployments (GitRepository + Flux Kustomization). All use `--ref master` and [kubeflow/manifests](https://github.com/kubeflow/manifests). See [docs/KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) for full context.

| App | Version | Command |
|-----|---------|---------|
| **Kubeflow Pipelines** | 2.15.0 | `./catalog-workflow.sh add-app --appname kubeflow-pipelines --version 2.15.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user --ref master --force` |
| **Jupyter Notebook Controller** | 1.10.0 | `./catalog-workflow.sh add-app --appname jupyter-notebook-controller --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/jupyter/notebook-controller/upstream/overlays/kubeflow --ref master --force` |
| **Training Operator** | 1.9.2 | `./catalog-workflow.sh add-app --appname training-operator --version 1.9.2 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/training-operator/upstream/overlays/kubeflow --ref master --force` |
| **Kubeflow Model Registry** | 0.3.6 | `./catalog-workflow.sh add-app --appname kubeflow-model-registry --version 0.3.6 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/model-registry/upstream --ref master --force` |
| **TensorBoard Controller** | 1.10.0 | `./catalog-workflow.sh add-app --appname tensorboard-controller --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/tensorboard/tensorboard-controller/upstream/overlays/kubeflow --ref master --force` |
| **Kubeflow Central Dashboard** | 1.10.0 | `./catalog-workflow.sh add-app --appname kubeflow-central-dashboard --version 1.10.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/centraldashboard/overlays/oauth2-proxy --ref master --force` |
| **Katib** | 0.19.0 | (already in catalog; alternative path) `./applications/katib/upstream/installs/katib-with-kubeflow` — see [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md). |

See [docs/KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) for full Kubeflow context and more components.

---

## Applications Not Using add-application.sh

### Job-Based / Custom Install (manual structure)

| App | Version | Notes |
|-----|---------|-------|
| **kagent** | 0.1.0 | Job-based install via `kubectl apply` of install.yaml. Copy structure from `applications/kagent/0.1.0/`. No OCI Helm chart. |

### Custom / GitOps Manifests (manual structure)

| App | Version | Notes |
|-----|---------|-------|
| **letsencrypt-clusterissuer** | 1.0.0 | Custom Kustomization with clusterissuers.yaml, cm.yaml. Depends on cert-manager. No upstream Helm chart. |

### Private / Custom Charts (manual or custom OCI)

| App | Version | OCI Path | Notes |
|-----|---------|----------|-------|
| **dm-nkp-gitops-custom-app** | 0.1.0 | `oci://ghcr.io/deepak-muley/dm-nkp-gitops-custom-app/prod/...` | Custom app; chart built and pushed from own repo. |
| **dm-nkp-gitops-a2a-server** | 0.2.0 | `oci://ghcr.io/deepak-muley/charts/dm-nkp-gitops-a2a-server/...` | Custom A2A server; chart in private OCI. |

---

## Post-add Steps

1. **Validate**: `./validate.sh`
2. **Update metadata.yaml**: categories, description, supportLink
3. **For agentgateway**: Ensure `ref.tag` matches available chart tag (e.g. `v2.2.0-main`)

---

## Deployment Order (Full Stack)

See [ARCHITECTURE.md](ARCHITECTURE.md) for details:

1. cert-manager  
2. letsencrypt-clusterissuer (optional)  
3. traefik (optional)  
4. kube-prometheus-stack, loki, tempo  
5. opentelemetry-operator (recommended) or opentelemetry-collector  
6. dm-nkp-gitops-custom-app (or your application)
