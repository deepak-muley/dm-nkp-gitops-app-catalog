# dm-nkp-gitops-app-catalog

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