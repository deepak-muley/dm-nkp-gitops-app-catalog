# Catalog Apps Guide

This guide describes each application in this NKP catalog: **what it is**, **what it does**, **whether it has a UI/dashboard**, **prerequisites**, **how to add it to the catalog**, **how to deploy it**, and **how to use it**. Use it to learn and onboard teams.

For add-app commands and versions, see [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md).

---

## Table of Contents

### AI / ML
- [vLLM](#vllm) · [KServe](#kserve) · [KServe CRD](#kserve-crd) · [Open WebUI](#open-webui) · [LocalAI](#local-ai) · [Weaviate](#weaviate) · [Katib](#katib) · [Kubeflow Pipelines](#kubeflow-pipelines) · [Jupyter Notebook Controller](#jupyter-notebook-controller) · [Training Operator](#training-operator) · [Kubeflow Model Registry](#kubeflow-model-registry) · [TensorBoard Controller](#tensorboard-controller) · [Kubeflow Central Dashboard](#kubeflow-central-dashboard) · [Spark Operator](#spark-operator) · [Agentgateway](#agentgateway) · [Slurm / Slurm Operator](#slurm--slurm-operator)

### Observability & Cost
- [Kube Prometheus Stack](#kube-prometheus-stack) · [Loki](#loki) · [Tempo](#tempo) · [OpenCost / Centralized OpenCost](#opencost--centralized-opencost) · [ClickStack](#clickstack) · [OpenTelemetry Collector](#opentelemetry-collector)

### Security & Policy
- [Cert Manager](#cert-manager) · [Kyverno](#kyverno) · [Kubescape Operator](#kubescape-operator) · [Vault](#vault) · [OAuth2 Proxy](#oauth2-proxy) · [Lets Encrypt ClusterIssuer](#lets-encrypt-clusterissuer)

### Infrastructure & Platform
- [Traefik](#traefik) · [KRO](#kro) · [Karmada Operator](#karmada-operator) · [Vertical Pod Autoscaler](#vertical-pod-autoscaler) · [Podinfo](#podinfo)

### Custom / Internal
- [Kagent](#kagent) · [dm-nkp-gitops-custom-app](#dm-nkp-gitops-custom-app) · [dm-nkp-gitops-a2a-server](#dm-nkp-gitops-a2a-server)

---

## AI / ML

### vLLM

| | |
|---|---|
| **What it is** | High-throughput, memory-efficient inference server for large language models (LLMs). |
| **What it does** | Serves open-source LLMs (Llama, Mistral, Qwen, etc.) with an **OpenAI-compatible API** (chat completions, embeddings). Uses PagedAttention and continuous batching for high throughput. |
| **UI / Dashboard** | No built-in UI. Use [Open WebUI](#open-webui) from this catalog for a chat interface, or call the API from code. |
| **Prerequisites** | Kubernetes 1.28+, NVIDIA GPU nodes with [NVIDIA Device Plugin](https://github.com/NVIDIA/k8s-device-plugin). Optional: Hugging Face token for gated models. |
| **How to use** | Deploy from catalog → set release name/namespace → configure model (e.g. `model.name`, `hfToken`). Port-forward or Ingress to the service, then use OpenAI client with `base_url` pointing to vLLM. |

**Full guide:** [VLLM-CATALOG.md](VLLM-CATALOG.md) — API endpoints, Helm values, Python/cURL examples, health check.

**Add to catalog:** See [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md) (Helm repo → OCI).

---

### KServe

| | |
|---|---|
| **What it is** | Kubernetes-native model serving layer. Turns trained models into scalable inference services behind a single **InferenceService** API. |
| **What it does** | Deploys and serves ML models (PyTorch, TensorFlow, scikit-learn, XGBoost, vLLM, custom) with serverless scaling to zero, canary/blue-green rollouts, and multi-framework support. Integrates with Kubeflow, MLflow, S3/GCS. |
| **UI / Dashboard** | No built-in UI. Use [Kubeflow Central Dashboard](#kubeflow-central-dashboard) to manage Kubeflow components; inference is via HTTP/gRPC. |
| **Prerequisites** | Kubernetes 1.28+. Install **KServe CRD** (same version) first. Optional: Knative, Istio, or Ingress for routing. |
| **How to use** | Deploy KServe CRD then KServe from catalog. Create `InferenceService` CRs (YAML or SDK) pointing to your model storage. Call the inference URL (predict, explain) or use the v2 OpenAPI contract. |

**References:** [KServe docs](https://kserve.github.io/website/) · [GitHub](https://github.com/kserve/kserve)

**Add to catalog:** OCI — see [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md). Set `ref.tag` to `v0.16.0` (or matching chart tag).

---

### KServe CRD

| | |
|---|---|
| **What it is** | Custom Resource Definitions for KServe (InferenceService, etc.). |
| **What it does** | Registers the KServe CRDs on the cluster. Required before installing KServe. |
| **UI / Dashboard** | No. |
| **How to use** | Install this app first, then install KServe (same version). |

---

### Open WebUI

| | |
|---|---|
| **What it is** | Web-based chat UI for any **OpenAI-compatible** API (vLLM, Ollama, OpenAI, etc.). |
| **What it does** | Provides a chat interface so users can talk to LLMs without writing code. Supports multiple models, conversations, and optional auth. Configure the backend URL to point to your vLLM (or other) service. |
| **UI / Dashboard** | **Yes.** Web UI for chat, model selection, and settings. |
| **Prerequisites** | An OpenAI-compatible backend (e.g. vLLM from this catalog) reachable from the cluster. |
| **How to use** | Deploy Open WebUI from catalog. Set Helm value for the API base URL (e.g. your vLLM service URL). Expose via Ingress or port-forward; open the UI in a browser and start chatting. |

**References:** [Open WebUI docs](https://docs.openwebui.com/) · [GitHub](https://github.com/open-webui/open-webui)

**Add to catalog:** [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md) — Helm repo → OCI.

---

### LocalAI

| | |
|---|---|
| **What it is** | OpenAI-compatible API server for running models **locally** (CPU or small GPU). |
| **What it does** | Serves local models (llama.cpp, etc.) with the same API shape as OpenAI. Good for dev, edge, or lightweight inference without heavy GPU. |
| **UI / Dashboard** | No built-in UI. Use [Open WebUI](#open-webui) or any OpenAI-compatible client. |
| **Prerequisites** | Kubernetes cluster; optional GPU. |
| **How to use** | Deploy from catalog, configure model path or download. Port-forward or Ingress, then use OpenAI client with `base_url` or point Open WebUI at it. |

**References:** [LocalAI](https://localai.io/) · [Helm](https://go-skynet.github.io/helm-charts/)

---

### Weaviate

| | |
|---|---|
| **What it is** | Vector database for embeddings and similarity search. |
| **What it does** | Stores and queries vector embeddings (e.g. from LLMs); supports hybrid search (vector + keyword). Used for RAG (retrieval-augmented generation), semantic search, and recommendation. |
| **UI / Dashboard** | Optional Weaviate Console (separate) for schema and data browsing; no UI in the Helm chart by default. |
| **Prerequisites** | Kubernetes; persistent storage. |
| **How to use** | Deploy from catalog. Connect from your app or LangChain/RAGFlow using Weaviate client; create classes (schemas), add vectors, query. |

**References:** [Weaviate docs](https://weaviate.io/developers/weaviate) · [Helm](https://weaviate.github.io/weaviate-helm/)

---

### Katib

| | |
|---|---|
| **What it is** | Kubernetes-native **hyperparameter tuning** and neural architecture search (NAS). |
| **What it does** | Runs automated tuning experiments (grid, random, Bayesian, etc.) for ML training jobs. Uses CRDs (Experiment, Trial). Integrates with Kubeflow Pipelines and training operators. |
| **UI / Dashboard** | **Yes.** Katib UI shows experiments and trials; often accessed via [Kubeflow Central Dashboard](#kubeflow-central-dashboard). |
| **Prerequisites** | Kubernetes; optional GPU for training jobs. |
| **How to use** | Deploy from catalog (Kustomize from kubeflow/manifests). Create `Experiment` CRs or use the UI to define search space and objective; Katib runs trials and finds better hyperparameters. |

**References:** [Katib docs](https://www.kubeflow.org/docs/components/hyperparameter-tuning/) · [GitHub](https://github.com/kubeflow/katib)

---

### Kubeflow Pipelines

| | |
|---|---|
| **What it is** | Platform for building and running **ML pipelines** (train → tune → serve) on Kubernetes. |
| **What it does** | Runs multi-step pipelines with experiments, artifact tracking, and recurring runs. Integrates with Katib, KServe, and Training Operator. |
| **UI / Dashboard** | **Yes.** Pipelines UI for designing pipelines, viewing runs, and comparing experiments. Often via [Kubeflow Central Dashboard](#kubeflow-central-dashboard). |
| **Prerequisites** | Kubernetes; cert-manager; optionally full Kubeflow stack for Central Dashboard. |
| **How to use** | Deploy from catalog (Kustomize). Access Pipelines UI, create pipelines via SDK or UI, trigger runs and view artifacts. |

**References:** [Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/) · [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md)

---

### Jupyter Notebook Controller

| | |
|---|---|
| **What it is** | Controller that manages **Jupyter notebooks** as Kubernetes resources (Notebook CRD). |
| **What it does** | Lets users spawn Jupyter notebooks on the cluster with configurable CPU/memory/GPU. Part of Kubeflow; integrates with Central Dashboard. |
| **UI / Dashboard** | **Yes.** Each notebook has its own JupyterLab/Notebook UI. Access via Central Dashboard or direct URL. |
| **Prerequisites** | Kubernetes; part of Kubeflow ecosystem. |
| **How to use** | Deploy from catalog. Create `Notebook` CRs or use Central Dashboard to start notebooks; connect to the notebook URL and code. |

**References:** [Kubeflow Notebooks](https://www.kubeflow.org/docs/components/notebooks/)

---

### Training Operator

| | |
|---|---|
| **What it is** | Kubernetes operator for **distributed ML training** (PyTorchJob, TFJob, XGBoostJob, etc.). |
| **What it does** | Runs distributed training jobs via CRDs; supports multi-node and GPU. Works with Katib and Pipelines. |
| **UI / Dashboard** | No built-in UI. View jobs via `kubectl` or Central Dashboard. |
| **Prerequisites** | Kubernetes; GPU nodes for GPU training. |
| **How to use** | Deploy from catalog. Create PyTorchJob/TFJob CRs (YAML or SDK); monitor with kubectl or Pipelines. |

**References:** [Training Operator](https://github.com/kubeflow/training-operator)

---

### Kubeflow Model Registry

| | |
|---|---|
| **What it is** | Central **model versioning and lineage** for ML models. |
| **What it does** | Stores model metadata and versions; links to pipeline runs and serving. Used with Pipelines and KServe. |
| **UI / Dashboard** | Depends on deployment; Model Registry may expose a UI or API-only. |
| **Prerequisites** | Kubernetes; database (chart/overlay may include it). |
| **How to use** | Deploy from catalog (uses standalone [kubeflow/model-registry](https://github.com/kubeflow/model-registry) repo, overlay `manifests/kustomize/overlays/db`). Register models via API or integrate with Pipelines. |

**References:** [Model Registry](https://github.com/kubeflow/model-registry) · [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md)

---

### TensorBoard Controller

| | |
|---|---|
| **What it is** | Controller that manages **TensorBoard** instances for visualizing training runs. |
| **What it does** | Spawns TensorBoard pods that read event logs (metrics, graphs, histograms) from training jobs. |
| **UI / Dashboard** | **Yes.** Each TensorBoard instance is a web UI. |
| **Prerequisites** | Kubernetes; training jobs that write TensorBoard events. |
| **How to use** | Deploy from catalog. Create TensorBoard CRs pointing to PVC or object storage with events; open the TensorBoard URL. |

**References:** [Kubeflow TensorBoard](https://www.kubeflow.org/docs/components/tensorboard/)

---

### Kubeflow Central Dashboard

| | |
|---|---|
| **What it is** | **Single entry-point UI** for Kubeflow (Pipelines, Notebooks, Katib, etc.). |
| **What it does** | Provides a central web UI to discover and launch notebooks, pipelines, experiments, and other Kubeflow apps. This catalog overlay uses oauth2-proxy for auth. |
| **UI / Dashboard** | **Yes.** The dashboard itself is the UI. |
| **Prerequisites** | Kubernetes; other Kubeflow components (Pipelines, Jupyter, etc.) for full value. |
| **How to use** | Deploy from catalog. Expose via Ingress; log in (oauth2-proxy) and use the dashboard to navigate to Pipelines, Notebooks, Katib, etc. |

**References:** [Central Dashboard](https://www.kubeflow.org/docs/components/central-dash/)

---

### Spark Operator

| | |
|---|---|
| **What it is** | Kubernetes operator for running **Apache Spark** applications. |
| **What it does** | Submits and manages Spark jobs (batch, streaming) as Kubernetes resources. Used for data processing and ML pipelines. |
| **UI / Dashboard** | Spark driver UIs (per application); no single central dashboard in the operator. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog. Create `SparkApplication` CRs; monitor via kubectl or Spark driver UI. |

**References:** [Spark on K8s](https://spark.apache.org/docs/latest/running-on-kubernetes.html) · [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md)

---

### Agentgateway

| | |
|---|---|
| **What it is** | AI-focused **gateway** (from kgateway.dev) for security, observability, and traffic management in front of LLM backends. |
| **What it does** | Sits between AI agents and LLM providers (e.g. vLLM, OpenAI); adds policy enforcement, logging, and rate limiting. Works with kagent for agent governance. |
| **UI / Dashboard** | Depends on kgateway.dev offering; check [kgateway.dev](https://kgateway.dev). |
| **Prerequisites** | Kubernetes; Gateway API CRDs. |
| **How to use** | Deploy in front of vLLM or other LLM APIs; route agent traffic through the gateway. |

**References:** [kgateway.dev](https://kgateway.dev)

---

### Slurm / Slurm Operator

| | |
|---|---|
| **What it is** | **Slurm** workload manager on Kubernetes via the Slinky **Slurm Operator** (SchedMD). |
| **What it does** | Runs Slurm clusters on K8s for HPC and large-scale ML training. Combines Slurm scheduling with Kubernetes orchestration. Install order: slurm-operator-crds → slurm-operator → slurm. |
| **UI / Dashboard** | Slurm has command-line tools (`squeue`, `sinfo`); optional web dashboards (e.g. Slurm Web) are separate. |
| **Prerequisites** | Kubernetes; cert-manager. |
| **How to use** | Deploy CRDs, then operator, then slurm chart. Submit jobs with `sbatch`/`srun` on login nodes. |

**References:** [Slinky Slurm Operator](https://slinky.schedmd.com/projects/slurm-operator) · [AI-APPS-RECOMMENDATIONS.md](AI-APPS-RECOMMENDATIONS.md)

---

## Observability & Cost

### Kube Prometheus Stack

| | |
|---|---|
| **What it is** | **Prometheus** + **Grafana** + Alertmanager + node/postgres etc. exporters in one Helm chart. |
| **What it does** | Collects metrics, stores them in Prometheus, visualizes in Grafana, and sends alerts. De facto standard for Kubernetes monitoring. |
| **UI / Dashboard** | **Yes.** Grafana (dashboards, explore); Prometheus (query UI); Alertmanager (alerts). |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog. Access Grafana (default admin credentials in values); import dashboards; add Prometheus datasource. Configure Alertmanager for alerts. |

**References:** [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

---

### Loki

| | |
|---|---|
| **What it is** | Log aggregation system (like Prometheus but for logs). |
| **What it does** | Ingests, indexes, and queries log streams; often used with Grafana for log exploration. |
| **UI / Dashboard** | No standalone UI; use **Grafana** with Loki as datasource. |
| **Prerequisites** | Kubernetes; optional object storage for retention. |
| **How to use** | Deploy from catalog. In Grafana, add Loki datasource and use Logs or Explore to query. |

**References:** [Loki](https://grafana.com/oss/loki/)

---

### Tempo

| | |
|---|---|
| **What it is** | Distributed **tracing** backend. |
| **What it does** | Stores and queries traces (OpenTelemetry, Jaeger); integrates with Grafana for trace visualization. |
| **UI / Dashboard** | Use **Grafana** with Tempo datasource (Trace view). |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog. Send traces from apps (OTLP or Jaeger); in Grafana add Tempo datasource and use Trace view. |

**References:** [Tempo](https://grafana.com/oss/tempo/)

---

### OpenCost / Centralized OpenCost

| | |
|---|---|
| **What it is** | **OpenCost** — open-source cost monitoring and allocation for Kubernetes. **Centralized OpenCost** runs the backend on the management cluster and collects cost from workload clusters. |
| **What it does** | Allocates and reports cost by namespace, deployment, label; optional UI for dashboards. |
| **UI / Dashboard** | **Yes.** OpenCost UI for cost dashboards and reports. |
| **Prerequisites** | Kubernetes; Prometheus (for metrics). |
| **How to use** | Deploy OpenCost (per cluster) or Centralized OpenCost (management cluster). Point to Prometheus; open UI for cost views. |

**Full guide:** [OPENCOST-MULTICLUSTER.md](OPENCOST-MULTICLUSTER.md)

**References:** [OpenCost](https://www.opencost.io/)

---

### ClickStack

| | |
|---|---|
| **What it is** | Observability platform on **ClickHouse**: logs, traces, metrics, sessions in one stack. |
| **What it does** | Unified storage and querying; OpenTelemetry collector; HyperDX UI for search, dashboards, alerting. |
| **UI / Dashboard** | **Yes.** ClickStack UI (HyperDX) for logs, traces, dashboards. |
| **Prerequisites** | Kubernetes; storage. |
| **How to use** | Deploy from catalog; ingest via OTLP or collector. Use the UI for search and dashboards. |

**Full guide:** [CLICKSTACK-CATALOG.md](CLICKSTACK-CATALOG.md)

---

### OpenTelemetry Collector

| | |
|---|---|
| **What it is** | **OTLP** collector: receives, processes, and exports traces, metrics, logs. |
| **What it does** | Central place to receive OpenTelemetry data and forward to backends (Prometheus, Loki, Tempo, etc.). |
| **UI / Dashboard** | No. |
| **Prerequisites** | Kubernetes; backends to export to. |
| **How to use** | Deploy from catalog; configure receivers and exporters. Point apps to the collector OTLP endpoint. |

**References:** [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)

---

## Security & Policy

### Cert Manager

| | |
|---|---|
| **What it is** | **Certificate management** for Kubernetes: issues and renews TLS certs (e.g. from Let's Encrypt). |
| **What it does** | Manages Certificate and Issuer/ClusterIssuer CRDs; automates cert lifecycle. |
| **UI / Dashboard** | No. Use `kubectl get certificate`. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog. Create Issuer/ClusterIssuer and Certificate CRs; cert-manager provisions certs; reference in Ingress. |

**References:** [cert-manager](https://cert-manager.io/)

---

### Kyverno

| | |
|---|---|
| **What it is** | **Policy engine** for Kubernetes (validate, mutate, generate policies). |
| **What it does** | Enforces policies as CRDs (ClusterPolicy); can block or mutate resources, generate configs. |
| **UI / Dashboard** | Optional Kyverno UI / reporting tools; core is policy-only. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog. Create ClusterPolicy CRs; policies apply to matching resources. |

**References:** [Kyverno](https://kyverno.io/)

---

### Kubescape Operator

| | |
|---|---|
| **What it is** | **Security scanning** for Kubernetes (compliance, vulnerabilities). |
| **What it does** | Scans clusters and workloads; reports risks and compliance. |
| **UI / Dashboard** | Kubescape has a SaaS/UI for results; operator may integrate. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog; configure scan targets and view results in Kubescape UI or CLI. |

**References:** [Kubescape](https://kubescape.io/)

---

### Vault

| | |
|---|---|
| **What it is** | **Secrets management** (HashiCorp Vault). |
| **What it does** | Stores and dynamically generates secrets; supports PKI, encryption, etc. |
| **UI / Dashboard** | **Yes.** Vault UI for secrets and config. |
| **Prerequisites** | Kubernetes; storage backend. |
| **How to use** | Deploy from catalog; unseal and configure. Use Vault UI or API/CLI to manage secrets; integrate apps via Vault Agent or CSI. |

**References:** [Vault](https://www.vaultproject.io/)

---

### OAuth2 Proxy

| | |
|---|---|
| **What it is** | **Reverse proxy** that adds OAuth2/OIDC authentication in front of services. |
| **What it does** | Protects upstream apps; users log in via IdP (Google, GitHub, etc.) before reaching the app. |
| **UI / Dashboard** | No; it’s a proxy (login page is minimal). |
| **Prerequisites** | Kubernetes; OAuth2/OIDC provider. |
| **How to use** | Deploy from catalog; put in front of Grafana, Central Dashboard, etc.; configure client ID/secret and upstream. |

**References:** [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/)

---

### Lets Encrypt ClusterIssuer

| | |
|---|---|
| **What it is** | **ClusterIssuer** resources for Let's Encrypt (staging and prod). |
| **What it does** | Lets cert-manager issue TLS certs from Let's Encrypt. |
| **UI / Dashboard** | No. |
| **Prerequisites** | Cert Manager installed. |
| **How to use** | Deploy after cert-manager; reference the ClusterIssuer in Certificate resources. |

---

## Infrastructure & Platform

### Traefik

| | |
|---|---|
| **What it is** | **Ingress controller** and reverse proxy. |
| **What it does** | Routes HTTP/HTTPS traffic to services; supports Ingress and IngressRoute CRDs, TLS, load balancing. |
| **UI / Dashboard** | **Yes.** Traefik dashboard (optional) for routes and services. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog; create Ingress or IngressRoute resources; enable dashboard in values if desired. |

**References:** [Traefik](https://doc.traefik.io/traefik/)

---

### KRO

| | |
|---|---|
| **What it is** | **Kubernetes Release Operator** — manages application releases and rollouts. |
| **What it does** | GitOps-friendly release and rollout automation. |
| **UI / Dashboard** | No. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Deploy from catalog; use KRO CRs for releases. |

**References:** [KRO](https://github.com/kubernetes-sigs/kro)

---

### Karmada Operator

| | |
|---|---|
| **What it is** | Operator for **Karmada** (multi-cluster Kubernetes). |
| **What it does** | Installs and manages Karmada control plane for multi-cluster scheduling and propagation. |
| **UI / Dashboard** | Karmada may have a web UI; check Karmada docs. |
| **Prerequisites** | Kubernetes (host cluster). |
| **How to use** | Deploy from catalog; register member clusters and deploy resources across clusters. |

**References:** [Karmada](https://karmada.io/)

---

### Vertical Pod Autoscaler

| | |
|---|---|
| **What it is** | **VPA** — adjusts pod CPU/memory requests and limits based on usage. |
| **What it does** | Recommends or applies vertical scaling; can avoid OOMs and over-provisioning. |
| **UI / Dashboard** | No. Use `kubectl describe vpa`. |
| **Prerequisites** | Kubernetes; metrics server. |
| **How to use** | Deploy from catalog; create VPA CRs targeting deployments; set update policy (Off/Initial/Recreate/Auto). |

**References:** [VPA](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler)

---

### Podinfo

| | |
|---|---|
| **What it is** | Small **demo app** for testing deployments and GitOps. |
| **What it does** | Serves a simple web page and API; used for canary, A/B, and rollout demos. |
| **UI / Dashboard** | **Yes.** Simple web UI showing pod info. |
| **Prerequisites** | None. |
| **How to use** | Deploy from catalog; access the service URL to see the demo UI. |

**References:** [podinfo](https://github.com/stefanprodan/podinfo)

---

## Custom / Internal

### Kagent

| | |
|---|---|
| **What it is** | **Agent** component from kgateway.dev for AI agent governance. |
| **What it does** | Works with Agentgateway to govern AI agent interactions with external LLMs. |
| **UI / Dashboard** | See kgateway.dev. |
| **Prerequisites** | Kubernetes. |
| **How to use** | Job-based install; deploy per catalog structure. See [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md). |

**References:** [kgateway.dev](https://kgateway.dev)

---

### dm-nkp-gitops-custom-app

| | |
|---|---|
| **What it is** | Placeholder / template **custom app** for this catalog. |
| **What it does** | Demonstrates structure for apps whose chart is built and pushed from your own repo. |
| **UI / Dashboard** | Depends on the app you build. |
| **How to use** | Replace with your own chart and metadata; use as reference for custom apps. |

---

### dm-nkp-gitops-a2a-server

| | |
|---|---|
| **What it is** | **A2A (Agent-to-Agent) server** — custom component for this environment. |
| **What it does** | Serves agent-to-agent protocol endpoints. |
| **UI / Dashboard** | Depends on implementation. |
| **How to use** | Deploy from catalog; configure per internal docs. |

---

## See also

- [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md) — add-app commands for every app
- [VLLM-CATALOG.md](VLLM-CATALOG.md) — detailed vLLM guide
- [CLICKSTACK-CATALOG.md](CLICKSTACK-CATALOG.md) — ClickStack guide
- [OPENCOST-MULTICLUSTER.md](OPENCOST-MULTICLUSTER.md) — OpenCost multi-cluster
- [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — Kubeflow components and paths
- [AI-APPS-RECOMMENDATIONS.md](AI-APPS-RECOMMENDATIONS.md) — which AI apps to add and why
