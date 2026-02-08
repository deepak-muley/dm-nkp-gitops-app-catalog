# vLLM in NKP Catalog

This document describes vLLM, how it is deployed via this catalog, and how to use it.

## What is vLLM?

**vLLM** is a high-throughput, memory-efficient inference and serving engine for large language models (LLMs). It provides:

- **OpenAI-compatible API** — Drop-in replacement for OpenAI's API; use the official `openai` Python client or any HTTP client pointing to your vLLM server
- **High throughput** — PagedAttention and continuous batching for 2–24× higher throughput than naive implementations
- **Low latency** — Optimized CUDA kernels and efficient memory management
- **Wide model support** — Llama, Mistral, Qwen, Phi, and many other open-source models from Hugging Face

### API Endpoints

vLLM exposes standard OpenAI endpoints:

| Endpoint | Description |
|----------|-------------|
| `/v1/chat/completions` | Chat completions (conversational) |
| `/v1/completions` | Text completions |
| `/v1/embeddings` | Embeddings |
| `/v1/models` | List available models |
| `/health` | Health check |

You can use the official OpenAI client by setting `base_url` to your vLLM service URL.

---

## Prerequisites

Before deploying vLLM via this catalog:

1. **Kubernetes cluster** — NKP or any Kubernetes 1.28+
2. **NVIDIA GPU nodes** — vLLM is designed for GPU inference. Ensure:
   - [NVIDIA Kubernetes Device Plugin](https://github.com/NVIDIA/k8s-device-plugin) is installed
   - Nodes have available `nvidia.com/gpu` resources
3. **Hugging Face token** (optional) — Required for gated models (e.g., Llama). Set `hfToken` in values.

---

## Adding vLLM to This Catalog

This catalog uses the [open-source-ai-dev/vllm-helm-chart](https://github.com/open-source-ai-dev/vllm-helm-chart) Helm chart. The chart is pulled from the Helm repo and pushed to OCI during add-app.

### Exact add-app Command

```bash
./catalog-workflow.sh add-app --appname vllm --version 0.1.1 \
  --helmrepo vllm/vllm \
  --ocipush oci://ghcr.io/deepak-muley/vllm \
  --helmrepo-url https://open-source-ai-dev.github.io/vllm-helm-chart
```

**Note:** You need `docker login ghcr.io` (or your OCI registry) with push access for the `--ocipush` path. Use `--force` to overwrite an existing app directory.

### Check for New Versions

```bash
./catalog-workflow.sh check-versions --appname vllm
```

---

## Deployment via NKP Catalog

When you deploy vLLM from the NKP catalog UI or via Flux/Kommander:

1. Choose the **vLLM** app and version **0.1.1**
2. Set **release name** and **namespace** (e.g., `vllm-llama`, `vllm`)
3. Configure values (see below)

### Key Helm Values

| Key | Default | Description |
|-----|---------|-------------|
| `model.organization` | `meta-llama` | Hugging Face org hosting the model |
| `model.name` | `Llama-3.2-1B` | Model name to serve |
| `model.temperature` | `0.95` | Sampling temperature |
| `model.contextLength` | `8192` | Max context length |
| `volumeSize` | `10` | PVC size in Gi for model cache |
| `hfToken` | `""` | Hugging Face token for gated models (use Secret in prod) |
| `image.name` | `vllm/vllm-openai` | Container image |
| `image.tag` | `""` | Image tag (empty = chart default) |
| `service.name` | `vllm-llama` | Kubernetes Service name (do not use `vllm` — see [vLLM docs](https://docs.vllm.ai)) |
| `ingress.enabled` | `false` | Enable Ingress for external access |

### Example ConfigMap Override

To serve a different model (e.g., Mistral), override values in the catalog's `cm.yaml` or via NKP UI:

```yaml
model:
  organization: "mistralai"
  name: "Mistral-7B-Instruct-v0.3"
  temperature: "0.7"
  contextLength: "32768"
volumeSize: "20"
hfToken: ""  # Set via Secret in production
```

---

## Using vLLM

### 1. Get the Service URL

After deployment, vLLM is exposed as a ClusterIP Service. To access it:

```bash
# Port-forward for local testing
kubectl port-forward svc/vllm-llama 8000:8000 -n vllm
```

Or expose via Ingress/Route if configured.

### 2. Chat Completions (OpenAI Client)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8000/v1",
    api_key="not-needed"  # vLLM doesn't require auth by default
)

response = client.chat.completions.create(
    model="Llama-3.2-1B",  # Must match --served-model-name
    messages=[{"role": "user", "content": "Hello, how are you?"}],
    max_tokens=100
)
print(response.choices[0].message.content)
```

### 3. cURL

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Llama-3.2-1B",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

### 4. Health Check

```bash
curl http://localhost:8000/health
```

---

## Official vLLM vs This Chart

The vLLM project also provides an [official Helm chart](https://docs.vllm.ai/en/latest/deployment/frameworks/helm/) in `examples/online_serving/chart-helm`. That chart:

- Supports S3 model download
- Uses different value structure
- Is not published to a Helm repo or OCI

This catalog uses the **open-source-ai-dev** community chart, which is simpler and published to a Helm repo. It pulls models from Hugging Face at startup.

---

## References

- [vLLM Documentation](https://docs.vllm.ai/)
- [vLLM GitHub](https://github.com/vllm-project/vllm)
- [OpenAI-Compatible Server](https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html)
- [vLLM Helm Chart (community)](https://github.com/open-source-ai-dev/vllm-helm-chart)
- [vLLM Helm (official)](https://docs.vllm.ai/en/latest/deployment/frameworks/helm/)
