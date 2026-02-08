# AI Apps Recommendations for Company Dev Catalog

Recommendations for which AI/ML applications to expose to developers via this NKP catalog. Use this to decide what to add and how to describe the stack to teams.

---

## What You Already Have (AI Stack)

| App | Purpose | Good for devs? |
|-----|---------|----------------|
| **vLLM** | LLM inference with OpenAI-compatible API | ✅ Yes — primary way to run Llama, Mistral, Qwen, etc. on GPU |
| **KServe** | Model serving (PyTorch, TensorFlow, SKLearn, XGBoost, vLLM) | ✅ Yes — production inference, canary, scale-to-zero |
| **KServe CRD** | CRDs for InferenceService | ✅ Yes — required if using KServe |
| **Katib** | Hyperparameter tuning & neural architecture search | ✅ Yes — ML engineers doing training optimization |
| **Spark Operator** | Run Apache Spark on Kubernetes | ✅ Yes — data prep and ML pipelines |
| **Agentgateway** | AI gateway (security, observability, traffic) for LLM backends | ✅ Yes — platform team; put in front of vLLM/APIs |

These form a solid base: **inference (vLLM + KServe)**, **training tuning (Katib)**, **data/ML jobs (Spark)**, and **governance (Agentgateway)**.

---

## Must-Use AI Apps — Recommended Additions

### Tier 1 — Strongly recommended for a broad dev-facing catalog

| App | Why | How to add |
|-----|-----|------------|
| **Kubeflow Pipelines** | End-to-end ML pipelines (train → tune → serve), experiments, recurring runs. Devs expect “run my pipeline from the catalog.” | Kustomize: see [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md). Example: `./catalog-workflow.sh add-app --appname kubeflow-pipelines --version 2.15.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user --ref master` |
| **Jupyter Notebook Controller** | Notebooks on Kubernetes; standard for data scientists and ML devs. | Kustomize: [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — `jupyter-notebook-controller` |
| **Training Operator** (or **Trainer** v2) | Distributed training (PyTorch, TensorFlow, etc.) on K8s. Complements Katib and Pipelines. | Kustomize: [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — `training-operator` or `trainer` |
| **Kubeflow Model Registry** | Central model versioning and lineage; ties into Pipelines and KServe. | Kustomize: [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — `kubeflow-model-registry` |

### Tier 2 — High value, add when use cases appear

| App | Why | How to add |
|-----|-----|------------|
| **TensorBoard Controller** | Visualization for training runs; pairs with Training Operator and Pipelines. | Kustomize: [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — `tensorboard-controller` |
| **Kubeflow Central Dashboard** | Single UI for Pipelines, Notebooks, Katib, etc. Improves discoverability for devs. | Kustomize: [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) — `kubeflow-central-dashboard` (choose overlay that matches your ingress/auth) |
| **Open WebUI** | Chat UI on top of OpenAI-compatible APIs (e.g. vLLM). Lets devs and product try LLMs without writing code. | Helm: `helm repo add open-webui https://helm.openwebui.com/` — add via `add-app --helmrepo open-webui/open-webui --ocipush oci://<your-registry>/open-webui --helmrepo-url https://helm.openwebui.com/` |

### Tier 3 — Optional / advanced

| App | Why |
|-----|-----|
| **Ray (KubeRay)** | Distributed training and serving; add if teams need Ray-native workloads. |
| **MLflow** | Experiment tracking and model registry; add if you standardize on MLflow instead of or alongside Kubeflow Model Registry. |
| **Ollama** | Simple local LLM runtime (CPU/small GPU). Consider only if you want a lightweight option alongside vLLM. |

---

## Open-source AI apps (Slurm, LangChain, RAG, and more)

Beyond Kubeflow and vLLM, these open-source projects have Helm charts or K8s operators and fit a company AI catalog. Add via `add-app --helmrepo` + `--ocipush` (then push to your OCI registry) or `--ocirepo` when the chart is already in OCI.

### HPC / job scheduling — Slurm

| App | What it is | Why consider | How to add |
|-----|------------|--------------|------------|
| **Slurm Operator** (Slinky) | SchedMD’s official Kubernetes operator for running Slurm clusters on K8s. CRDs for cluster lifecycle, GPU support, accounting. | Teams with HPC/ML training workloads that already use Slurm, or who want Slurm-style job scheduling and resource efficiency on K8s. Complements K8s-native training (e.g. Training Operator). | **OCI:** `helm install slurm-operator oci://ghcr.io/slinkyproject/charts/slurm-operator` — add to catalog with `--ocirepo oci://ghcr.io/slinkyproject/charts/slurm-operator`. Also install **slurm-operator-crds** and **slurm** (cluster) from same org. Requires cert-manager. [Slinky docs](https://slinky.schedmd.com/projects/slurm-operator). |
| **Soperator** (Nebius) | Open-source K8s operator for Slurm by Nebius. GPU-ready, fault tolerance via K8s. | Alternative if you prefer Nebius’ layout or tooling. | See [Nebius Soperator](https://nebius.com/blog/posts/introducing-soperator). |

**Note:** Slurm vs Kubernetes is a common choice for large-scale training; running Slurm on K8s (via an operator) lets you share GPU capacity and use K8s for orchestration while keeping Slurm semantics.

### LangChain ecosystem

LangChain is a framework for building LLM apps (chains, agents, RAG). These are **deployable** components that fit a catalog:

| App | What it is | Why consider | How to add |
|-----|------------|--------------|------------|
| **LangSmith** | Observability, tracing, and evaluation for LangChain apps. | Teams using LangChain/LangServe who want tracing and eval in their own infra. | Helm: `helm repo add langchain https://langchain-ai.github.io/helm/` — charts: `langsmith`, `langsmith-observability`. Add via `add-app --helmrepo langchain/langsmith --ocipush oci://<registry>/langsmith --helmrepo-url https://langchain-ai.github.io/helm/`. |
| **LangGraph Cloud** | API server + PostgreSQL for LangGraph (stateful agent workflows). | Teams building multi-step/agent apps with LangGraph who want a hosted control plane. | Same Helm repo — chart `langgraph-cloud`. |
| **LangGraph DataPlane** | Operator + CRDs for managing LangGraph deployments on K8s. | When you want to manage LangGraph deployments as K8s resources. | Same Helm repo — chart `langgraph-dataplane`. |

**LangServe** (deploying LangChain runnables as APIs) is typically deployed as your own app image on top of KServe or a normal Deployment; the **langchain-ai/helm** repo focuses on LangSmith, LangGraph Cloud, and LangGraph DataPlane. Repo: [langchain-ai/helm](https://github.com/langchain-ai/helm), add repo: `https://langchain-ai.github.io/helm/`.

### Chat UIs and local inference

| App | What it is | Why consider | How to add |
|-----|------------|--------------|------------|
| **Open WebUI** | Chat UI (formerly Ollama WebUI) for OpenAI-compatible backends (vLLM, Ollama, etc.). | Let devs and product use LLMs via a UI without writing code. Pairs with vLLM. | Helm: `https://helm.openwebui.com/` — chart `open-webui/open-webui`. Use `add-app --helmrepo` + `--ocipush`. |
| **LocalAI** | OpenAI-compatible API server for local models (CPU/small GPU). Many model backends (llama.cpp, etc.). | Lightweight inference without heavy GPU; good for dev or edge. | Helm: `helm repo add go-skynet https://go-skynet.github.io/helm-charts/` — chart `go-skynet/local-ai`. Or OCI: TrueCharts `oci://oci.trueforge.org/truecharts/local-ai`. |
| **Anything-LLM** | All-in-one LLM workspace (chat, documents, embeddings). | Single app for chat + document RAG in one place. | Community Helm: [la-cc/anything-llm-helm-chart](https://github.com/la-cc/anything-llm-helm-chart), Artifact Hub `anything-llm-helm-chart/anything-llm`. |

### RAG and document AI

| App | What it is | Why consider | How to add |
|-----|------------|--------------|------------|
| **RAGFlow** | RAG engine: document parsing, chunking, retrieval, LLM generation. | Teams standardizing on a full RAG pipeline (docs → retrieval → LLM). | Community Helm: [fzhan/ragflow-helm](https://github.com/fzhan/ragflow-helm) (Apache-2.0). Requires MySQL, Redis, optional MinIO/Elasticsearch. |

### Vector databases (for RAG and embeddings)

Used by LangChain, RAGFlow, and custom RAG stacks for storing and querying embeddings:

| App | What it is | Why consider | How to add |
|-----|------------|--------------|------------|
| **Weaviate** | Vector DB with hybrid search (vector + keyword), multi-tenancy. BSD-3. | Production RAG; strong search and scale. | Helm: `helm repo add weaviate https://weaviate.github.io/weaviate-helm/` — chart version 17.x. Use `add-app --helmrepo` + `--ocipush`. |
| **Qdrant** | Vector DB (Rust), fast queries. Apache 2.0. | Low-latency vector search; good for RAG and similarity search. | Deploy via container/StatefulSet or community operators; check Artifact Hub or [Qdrant docs](https://qdrant.tech/documentation/) for Helm/OCI. |
| **Chroma** | Lightweight, Python-native vector store. | Prototyping and dev; often embedded in app rather than separate cluster service. | Typically run as sidecar or in-app; less common as a standalone catalog app. |

### Summary — open-source additions to consider

- **Slurm** — Add **Slurm Operator** (Slinky) if you have HPC/ML training teams who want Slurm on K8s.
- **LangChain** — Add **LangSmith** (observability) and/or **LangGraph Cloud** / **LangGraph DataPlane** if teams use LangChain/LangServe/LangGraph.
- **Chat / UX** — Add **Open WebUI** (and optionally **LocalAI** or **Anything-LLM**) so non-engineers can use vLLM/backends via a UI.
- **RAG** — Add **RAGFlow** for full document RAG; add **Weaviate** (or **Qdrant**) as the vector store for RAG and LangChain apps.

After adding any of these: run `./catalog-workflow.sh validate`, set `metadata.yaml` (categories: `dm-nkp-gitops-app-catalog`, `ai-ml`, etc.), then build-push. For Helm-repo-based charts, add a `.catalog-source.yaml` so `check-versions` can recommend updates — see [CATALOG-SOURCE.md](CATALOG-SOURCE.md).

---

## Minimal “must-use” set for most companies

For a **minimal but complete** AI catalog that covers most dev needs:

1. **Keep and promote** (you already have): **vLLM**, **KServe** (+ **KServe CRD**), **Katib**, **Spark Operator**, **Agentgateway**.
2. **Add**:
   - **Kubeflow Pipelines** — so devs can run and schedule ML pipelines.
   - **Jupyter Notebook Controller** — so data scientists and ML devs can run notebooks on the cluster.
   - **Training Operator** or **Trainer** — so teams can run distributed training jobs.
   - **Kubeflow Model Registry** — so models are versioned and discoverable for KServe/Pipelines.

That gives you: **notebooks** → **training** → **tuning (Katib)** → **pipelines** → **model registry** → **inference (vLLM/KServe)** → **gateway (Agentgateway)**.

---

## Categories and metadata

For every AI app in this catalog, in `metadata.yaml` use categories such as:

- `dm-nkp-gitops-app-catalog`
- `ai-ml` and/or `artificial-intelligence`
- `infrastructure` where appropriate

See [APPLICATION-METADATA-FIELDS.md](APPLICATION-METADATA-FIELDS.md). This keeps AI apps easy to filter in the NKP UI.

---

## Dependencies and order

- **KServe** requires **KServe CRD** (same version).
- **Kubeflow Pipelines / Jupyter / Central Dashboard** may expect **cert-manager**, **Istio** (or another ingress), and optionally **Profiles + KFAM** if you go full Kubeflow. For “catalog-only” usage, start with Pipelines + Notebook Controller and add dependencies as needed; see [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md) for install order.
- **Agentgateway** works in front of vLLM or any OpenAI-compatible endpoint; no hard dependency on vLLM in catalog, but document that they are often used together.

---

## Quick reference — add-app commands (from this repo)

- **vLLM** (already added): Helm repo → OCI — see [ADD-APPLICATION-COMMANDS.md](ADD-APPLICATION-COMMANDS.md) and [VLLM-CATALOG.md](VLLM-CATALOG.md).
- **KServe / KServe CRD**: OCI — see [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md); set `ref.tag` to `v0.16.0` (or matching chart tag).
- **Katib, Pipelines, Jupyter, Training Operator, Model Registry, TensorBoard, Central Dashboard**: Kustomize via `./catalog-workflow.sh add-app --kustomize --gitrepo https://github.com/kubeflow/manifests --path <path> --ref master` — paths in [KUBEFLOW-CATALOG.md](KUBEFLOW-CATALOG.md).
- **Open-source AI (Slurm, LangChain, Open WebUI, Weaviate, etc.)**: See section [Open-source AI apps (Slurm, LangChain, RAG, and more)](#open-source-ai-apps-slurm-langchain-rag-and-more) above for Helm repo URLs and `add-app --helmrepo` + `--ocipush` examples. Slurm Operator is OCI: `oci://ghcr.io/slinkyproject/charts/slurm-operator`.

After adding any app: run `./catalog-workflow.sh validate`, update `metadata.yaml` (description, supportLink, categories), then build-push as needed.
