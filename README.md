# dm-nkp-gitops-app-catalog

## Applications in this Catalog

### Infrastructure Applications

| Application | Version | Category | Description |
|-------------|---------|----------|-------------|
| cert-manager | 1.19.2 | security, certificates | Certificate management controller |
| letsencrypt-clusterissuer | 1.0.0 | security, certificates | Let's Encrypt ACME ClusterIssuers |
| traefik | 38.0.2 | networking, ingress | Cloud-native edge router and ingress controller |

### Observability Stack (Latest Versions)

| Application | Version | Category | Description |
|-------------|---------|----------|-------------|
| kube-prometheus-stack | 80.13.3 | observability, monitoring | Prometheus + Grafana + Alertmanager monitoring stack |
| loki | 6.46.0 | observability, logging | Log aggregation system with OTLP support |
| tempo | 1.21.1 | observability, tracing | Distributed tracing backend |
| **opentelemetry-operator** | 0.140.0 | observability, operators | **Recommended** - Manages OTel Collectors via CRDs |
| opentelemetry-collector | 0.140.1 | observability, monitoring | Alternative - Standalone collector deployment |

### Demo Applications

| Application | Version | Category | Description |
|-------------|---------|----------|-------------|
| dm-nkp-gitops-custom-app | 0.1.0 | observability, demo | Demo app with OpenTelemetry telemetry (metrics, logs, traces) |
| podinfo | 6.9.4 | general | Demo/test application |

### Security Applications

| Application | Version | Category | Description |
|-------------|---------|----------|-------------|
| kyverno | 3.6.1 | security | Policy engine for Kubernetes |
| kubescape-operator | 1.29.12 | security | Security scanning and compliance |
| vault | 0.31.0 | security | Secrets management |
| oauth2-proxy | 8.0.2 | security | OAuth2 authentication proxy |

### Multi-Cluster & Platform

| Application | Version | Category | Description |
|-------------|---------|----------|-------------|
| karmada-operator | 1.16.0 | multi-cluster | Multi-cluster management |
| kro | 0.7.1 | general | Kubernetes Resource Orchestrator |

## Full Stack Architecture for dm-nkp-gitops-custom-app

The `dm-nkp-gitops-custom-app` can be deployed with the complete infrastructure stack:

```
                              ┌─────────────────────────────────┐
                              │          Internet               │
                              └───────────────┬─────────────────┘
                                              │
                              ┌───────────────▼─────────────────┐
                              │     cert-manager (1.19.2)       │
                              │  + letsencrypt-clusterissuer    │
                              │    (TLS Certificate Automation) │
                              └───────────────┬─────────────────┘
                                              │
                              ┌───────────────▼─────────────────┐
                              │      Traefik (38.0.2)           │
                              │    (Ingress Controller)         │
                              │  - TLS termination              │
                              │  - Automatic HTTPS              │
                              └───────────────┬─────────────────┘
                                              │
┌─────────────────────────────────────────────▼─────────────────────────────────────────────┐
│                              dm-nkp-gitops-custom-app (0.1.0)                             │
│                          (exports metrics, logs, traces via OTLP)                         │
└─────────────────────────────────────────────┬─────────────────────────────────────────────┘
                                              │ OTLP (gRPC :4317)
                                              ▼
┌───────────────────────────────────────────────────────────────────────────────────────────┐
│                    OpenTelemetry Operator (0.140.0) + Collector CR                        │
│              (manages collector via OpenTelemetryCollector CRD named "otel")              │
│                                                                                           │
│   ┌───────────────────────────────────────────────────────────────────────────────────┐   │
│   │                    otel-collector (deployed by operator)                          │   │
│   │                  (receives, processes, exports telemetry)                         │   │
│   └───────────────────────────────────────────────────────────────────────────────────┘   │
└─────────────┬───────────────────────────────┬───────────────────────────────┬─────────────┘
              │ Prometheus                    │ OTLP                          │ OTLP
              │ Remote Write                  │ HTTP                          │ gRPC
              ▼                               ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│  kube-prometheus-stack  │     │     Loki (6.46.0)       │     │    Tempo (1.21.1)       │
│       (80.13.3)         │     │                         │     │                         │
│  (Prometheus + Grafana) │     │    (Logs storage)       │     │   (Traces storage)      │
└───────────┬─────────────┘     └───────────┬─────────────┘     └───────────┬─────────────┘
            │                               │                               │
            └───────────────────────────────┼───────────────────────────────┘
                                            │
                                            ▼
                            ┌───────────────────────────────────┐
                            │      Grafana (included in         │
                            │    kube-prometheus-stack)         │
                            │   - Metrics dashboards            │
                            │   - Logs exploration              │
                            │   - Trace visualization           │
                            │   - Trace-to-logs correlation     │
                            └───────────────────────────────────┘
```

### Deployment Order

Deploy the full stack in this order:

#### Phase 1: Infrastructure (Required for Operator)
1. **cert-manager** - Certificate management controller (required for operator webhooks)
2. **letsencrypt-clusterissuer** - Let's Encrypt staging & production issuers (optional)
3. **traefik** - Ingress controller with TLS support (optional)

#### Phase 2: Observability Storage Backends (Required)
4. **kube-prometheus-stack** - Prometheus for metrics, Grafana for visualization
5. **loki** - Log aggregation with OTLP support
6. **tempo** - Distributed tracing backend

#### Phase 3: Telemetry Collection (Choose One)
7. **opentelemetry-operator** (Recommended for NKP) - Manages collectors via CRDs
   - After operator is running, apply the `OpenTelemetryCollector` CR named "otel"
   - This creates a service `otel-collector:4317` that the app connects to

   OR

7. **opentelemetry-collector** (Alternative) - Standalone collector deployment

#### Phase 4: Application
8. **dm-nkp-gitops-custom-app** - The demo application

### Quick Deploy Commands

```bash
# Infrastructure (cert-manager required for operator)
./add-application.sh --appname cert-manager --version 1.19.2 --ocirepo oci://quay.io/jetstack/charts/cert-manager
./add-application.sh --appname traefik --version 38.0.2 --ocirepo oci://ghcr.io/traefik/helm/traefik

# Observability Storage Backends
./add-application.sh --appname kube-prometheus-stack --version 80.13.3 --ocirepo oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
./add-application.sh --appname loki --version 6.46.0 --ocirepo oci://ghcr.io/grafana/helm-charts/loki
./add-application.sh --appname tempo --version 1.21.1 --ocirepo oci://ghcr.io/grafana/helm-charts/tempo

# Option A: OpenTelemetry Operator (RECOMMENDED for NKP)
./add-application.sh --appname opentelemetry-operator --version 0.140.0 --ocirepo oci://ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-operator
# Then apply the OpenTelemetryCollector CR:
# kubectl apply -f applications/opentelemetry-operator/0.140.0/helmrelease/otel-collector-cr.yaml

# Option B: Standalone OpenTelemetry Collector (alternative)
./add-application.sh --appname opentelemetry-collector --version 0.140.1 --ocirepo oci://ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-collector
```

### OpenTelemetry Operator vs Collector

| Feature | opentelemetry-operator | opentelemetry-collector |
|---------|------------------------|-------------------------|
| Deployment | Via `OpenTelemetryCollector` CRD | Direct Helm deployment |
| Management | Kubernetes-native, declarative | Helm values configuration |
| Auto-instrumentation | Supported (Java, Node, Python, .NET, Go) | Not supported |
| Multiple collectors | Easy via multiple CRs | Requires multiple Helm releases |
| NKP Integration | Recommended (matches [nkp-nutanix-product-catalog](https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog)) | Works but less integrated |
| cert-manager | Required for webhooks | Not required |

### dm-nkp-gitops-custom-app Compatibility

The [dm-nkp-gitops-custom-app](https://github.com/deepak-muley/dm-nkp-gitops-custom-app) connects to `otel-collector:4317` for OTLP telemetry export.

**With opentelemetry-operator:**
- The included `OpenTelemetryCollector` CR is named `otel`
- The operator creates a service named `otel-collector` (pattern: `<name>-collector`)
- This matches what the app expects - **fully compatible**

**With opentelemetry-collector:**
- The Helm chart creates a service named `opentelemetry-collector`
- You may need to update the app's `OTEL_EXPORTER_OTLP_ENDPOINT` to `opentelemetry-collector:4317`
- OR rename the Helm release to `otel` so the service becomes `otel-collector`

### Let's Encrypt Configuration

After deploying cert-manager and letsencrypt-clusterissuer, update the ACME email in:
`applications/letsencrypt-clusterissuer/1.0.0/helmrelease/clusterissuers.yaml`

Two ClusterIssuers are provided:
- `letsencrypt-staging` - For testing (certificates not trusted by browsers)
- `letsencrypt-prod` - For production (trusted certificates)

Example Ingress annotation for automatic certificates:
```yaml
annotations:
  cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

## Doc references
- https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-custom-apps-c.html

- https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-partner-catalog-in-nkp-c.html

- https://portal.nutanix.com/page/documents/details?targetId=Nutanix-Kubernetes-Platform-v2_16:top-workspace-app-metadata-c.html

In a fresh dir
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main) $ ls
README.md
</pre>

Generate app structure
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ nkp generate catalog-repository --apps=podinfo=6.9.4
Successfully initialized application layout for podinfo-6.9.4
Catalog layout generated at /Users/deepak.muley/go/src/github.com/deepak-muley/dm-nkp-gitops-app-catalog
</pre>

view dir structure
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ tree .
.
├── applications
│   └── podinfo
│       └── 6.9.4
│           ├── helmrelease
│           │   ├── cm.yaml
│           │   ├── helmrelease.yaml
│           │   └── kustomization.yaml
│           ├── helmrelease.yaml
│           ├── kustomization.yaml
│           └── metadata.yaml
└── README.md

5 directories, 7 files
</pre>

update the yamls and validate schema
<pre>
 deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ nkp validate catalog-repository --repo-dir=.
 ∅ Validating user-inputs.yaml for podinfo/6.9.4
 ✓ K8s [1.33.0]: Parsing resources
 ✓ K8s v1.33.0: Validating
APP      VERSION  METADATA  APP MANIFESTS  ERROR
podinfo  6.9.4    ✓         ✓
Validation completed in 2.5 seconds
</pre>

Create catalog bundle tgz from local structure
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ nkp create catalog-bundle --collection-tag v0.1.0
Bundling 1 application(s) (airgapped : false)
 ✓ Building OCI artifact dm-nkp-gitops-app-catalog/collection:v0.1.0
 ✓ Building OCI artifact dm-nkp-gitops-app-catalog/podinfo:6.9.4
Processing application podinfo/6.9.4
 ✓ K8s [1.33.0]: Parsing resources
 ✓ K8s v1.33.0: Validating
 ✓ Pulling requested images [====================================>2/2] (time elapsed 00s)
 ✓ Saving application bundle to /Users/deepak.muley/go/src/github.com/deepak-muley/dm-nkp-gitops-app-catalog/dm-nkp-gitops-app-catalog.tar

Run the following to push the artifact to your registry:

	nkp push bundle --bundle /Users/deepak.muley/go/src/github.com/deepak-muley/dm-nkp-gitops-app-catalog/dm-nkp-gitops-app-catalog.tar --to-registry <your-registry-url>


Run the following command to create catalog artifact(s) after pushing them:

	nkp create catalog-collection --url oci://<registry-url>/dm-nkp-gitops-app-catalog/collection --tag v0.1.0 --workspace kommander-workspace
</pre>

To push the catalog bundle to oci repo, Create a GitHub Personal Access Token (PAT) with the following scopes:
write:packages
read:packages
delete:packages (optional)
Go to: GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → Generate new token

Login
<pre>
echo "YOUR_GITHUB_PAT" | docker login ghcr.io -u "YOUR_USER_NAME" --password-stdin
or
docker login ghcr.io
or
echo "${GHCR_TOKEN}" | oras login --username "${GHCR_USER}" --password-stdin ghcr.io
</pre>

Push bundle to OCI repo
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ nkp push bundle --bundle /Users/deepak.muley/go/src/github.com/deepak-muley/dm-nkp-gitops-app-catalog/dm-nkp-gitops-app-catalog.tar --to-registry oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog --to-registry-username string --to-registry-password string
 ✓ Creating temporary directory
 ✓ Extracting bundle configs from "/Users/deepak.muley/go/src/github.com/deepak-muley/dm-nkp-gitops-app-catalog/dm-nkp-gitops-app-catalog.tar"
 ✓ Parsing image bundle config
 ✓ Starting temporary Docker registry
 ✓ Pushing bundled images [====================================>2/2] (time elapsed 04s)
</pre>

Verify if artifacts are pushed

You can see your uploaded packages under packages tab eg:https://github.com/deepak-muley?tab=packages
uploaded artifacts are private by default

you can also verify using
<pre>
deepak.muley:dm-nkp-gitops-app-catalog/ (main✗) $ oras discover ghcr.io/deepak-muley/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection:v0.1.0         [11:11:13]
ghcr.io/deepak-muley/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection@sha256:c36861d4a1378f081cafa6b70ef25a83ecdbb994de149d0d3c8c0086b1defaf0
</pre>

## Automated Build and Push Script

Use the `build-and-push.sh` script to automate validation, bundle creation, and pushing:

### Setup Credentials

First, set up your credentials as environment variables. You can use the helper script:

<pre>
cp setup-credentials.sh.example setup-credentials.sh
# Edit setup-credentials.sh with your credentials
source setup-credentials.sh
</pre>

Or export manually:
<pre>
export GHCR_USERNAME=deepak-muley
export GHCR_PASSWORD=your-github-pat
export MAKE_PUBLIC=false  # Set to "true" to make package public after pushing
</pre>

### Run the Script

<pre>
./build-and-push.sh v0.1.0
</pre>

The script will:
1. Validate the catalog repository
2. Create the catalog bundle with the specified tag
3. Push the bundle to GitHub Container Registry
4. Optionally make the package public (if MAKE_PUBLIC=true)

**Note:** Credentials are stored in `setup-credentials.sh` which is excluded from git via `.gitignore`. Never commit credentials to the repository.

Now import catalog bundle by creating a catalog collection in your NKP mgmt Cluster
<pre>
[dm-nkp-mgmt-1-admin@dm-nkp-mgmt-1|default] deepak.muley:nkp/ $ nkp create catalog-collection --url oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection --workspace dm-dev-workspace --tag v0.1.0 --dry-run -oyaml

apiVersion: source.toolkit.fluxcd.io/v1
kind: OCIRepository
metadata:
  creationTimestamp: null
  labels:
    catalog.nkp.nutanix.com/catalog-source-artifact: "true"
  name: dm-nkp-gitops-app-catalog-collection
  namespace: dm-dev-workspace
spec:
  interval: 6h0m0s
  ref:
    tag: v0.1.0
  timeout: 1m0s
  url: oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection
status: {}
</pre>

Note that currently NKP does not support HelmRepository in catalog bundle.
if you want to continue using catalog collection cli then you need to pull the helm chart tgz and push it as oci in your own oci repo.
following will download kubescape-operator-1.29.12.tgz
<pre>
helm repo add kubescape https://kubescape.github.io/helm-charts/
helm search repo kubescape/kubescape-operator --versions
helm pull kubescape/kubescape-operator
helm push kubescape-operator-1.29.12.tgz oci://ghcr.io/deepak-muley/kubescape-operator
</pre>

<pre>
helm repo add kyverno https://kyverno.github.io/kyverno/
helm pull kyverno/kyverno
helm push kyverno-3.6.1.tgz oci://ghcr.io/deepak-muley/kyverno
</pre>

<pre>
helm repo add hashicorp https://helm.releases.hashicorp.com
helm search repo hashicorp/vault
helm pull hashicorp/vault
helm push vault-0.31.0.tgz oci://ghcr.io/deepak-muley/vault
</pre>

<pre>
helm repo add karmada-charts https://raw.githubusercontent.com/karmada-io/karmada/master/charts
helm repo update karmada-charts
helm search repo karmada-charts/karmada --versions
helm search repo karmada-charts/karmada-operator --versions
helm pull karmada-charts/karmada-operator
helm push karmada-operator-chart-v1.16.0.tgz oci://ghcr.io/deepak-muley/karmada-operator
</pre>

Commands to add appliction structure for
kubescape
<pre>
nkp generate catalog-repository --apps=kubescape-operator=1.29.12
</pre>

vault
<pre>
nkp generate catalog-repository --apps=vault=0.31.0
</pre>

kyverno
<pre>
nkp generate catalog-repository --apps=kyverno=3.6.1
</pre>


## How to get values from helm chart
from oci repo helm chart
```bash
helm show values oci://registry-1.docker.io/bitnamicharts/harbor
```

from non oci repo helm chart
```bash
helm repo add harbor https://helm.goharbor.io
helm repo update
helm show values harbor/harbor
```

from already deployed release
```bash
helm get values <release-name> -n <namespace>
```

for kommander-applications
ref: https://github.com/mesosphere/kommander-applications/blob/main/applications/kube-prometheus-stack/78.4.0/helmrelease/kube-prometheus-stack.yaml 
```bash
helm show values oci://ghcr.io/mesosphere/charts/kube-prometheus-stack | wc -l
Pulled: ghcr.io/mesosphere/charts/kube-prometheus-stack:78.4.0
Digest: sha256:315be770eb7ed613ad4aebb4a07a0e50b79eb9f76f2fddaa00da27d894b890b7
    5465
```

ref: https://github.com/mesosphere/kommander-applications/blob/main/applications/nkp-insights/1.7.3/helmrelease/ocirepository.yaml
```bash
helm show values oci://ghcr.io/mesosphere/charts/nkp-insights | wc -l         
Pulled: ghcr.io/mesosphere/charts/nkp-insights:1.7.3
Digest: sha256:294c259a603e94f68f6f25b2bade0b69f3046ea5227f6703b87b10710c97e76b
     573
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/nutanix-ai/2.5.0/helmreleases/nai-operators.yaml
```bash
helm show values oci://ghcr.io/mesosphere/charts/nai-operators | wc -l
Pulled: ghcr.io/mesosphere/charts/nai-operators:2.5.0
Digest: sha256:f4e3d28d1333ffb8b9eb0354b76b567ed3f9ad41df53990ea38b16fa15245b2f
      88
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/nutanix-ai/2.5.0/helmreleases/nutanix-ai.yaml
```bash
helm show values oci://ghcr.io/mesosphere/charts/nai-core | wc -l
Pulled: ghcr.io/mesosphere/charts/nai-core:2.5.0
Digest: sha256:8fed4a68c1070c1dc960f2f2dfa10b3f7e57796c0c4826fea43c3c79b6104875
     593
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/ndk/2.0.0/release/ndk.yaml
```bash
helm show values oci://ghcr.io/mesosphere/charts/ndk | wc -l 
Pulled: ghcr.io/mesosphere/charts/ndk:2.0.0
Digest: sha256:429c80859a13f207e489b82efceb48830a43bba2ea1879d0dd1512f6ccf60e94
     272
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/envoy-gateway/1.5.0/helmrelease/envoy-gateway.yaml
```bash
helm show values oci://docker.io/envoyproxy/gateway-helm | wc -l
Pulled: docker.io/envoyproxy/gateway-helm:1.6.1
Digest: sha256:a89ce554ad10b951e521ac543687ce41616932b2b614121eeffda4c174d2c9dc
     145
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/kserve/0.15.0/helmrelease/kserve.yaml
```bash
helm show values oci://ghcr.io/kserve/charts/kserve-crd     
Error: unable to locate any tags in provided repository: oci://ghcr.io/kserve/charts/kserve-crd

helm show values oci://ghcr.io/kserve/charts/kserve    
Error: unable to locate any tags in provided repository: oci://ghcr.io/kserve/charts/kserve
```

ref: https://github.com/nutanix-cloud-native/nkp-nutanix-product-catalog/blob/release-2.x/applications/opentelemetry-operator/0.93.0/helmrelease/opentelemetry.yaml 
```bash
helm show values oci://ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-operator | wc -l
Pulled: ghcr.io/open-telemetry/opentelemetry-helm-charts/opentelemetry-operator:0.102.0
Digest: sha256:bb7b81fe9471479cd280bab46cad0c96c2127e4368bf4911feeb40202c363f75
     416
```