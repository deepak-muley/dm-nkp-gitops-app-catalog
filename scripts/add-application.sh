#!/bin/bash

# Script to add a new application to the NKP catalog repository
# Usage:
#   Helm (OCI): ./add-application.sh --appname <name> --version <version> --ocirepo <oci-path>
#   Helm (repo): ./add-application.sh --appname <name> --version <version> --helmrepo <repo/chart> --ocipush <oci-path>
#   Kustomize: ./add-application.sh --appname <name> --version <version> --kustomize --gitrepo <url> --path <path>
# Example: ./add-application.sh --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo
# Example: ./add-application.sh --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize variables
APPNAME=""
VERSION=""
OCIREPO=""
HELMREPO=""
OCIPUSH=""
HELMREPO_URL=""
FORCE=false
KUSTOMIZE=false
GITREPO=""
KUSTOMIZE_PATH=""
GITREF="main"
SKIP_VALIDATE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --appname)
            APPNAME="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --ocirepo)
            OCIREPO="$2"
            shift 2
            ;;
        --helmrepo)
            HELMREPO="$2"
            shift 2
            ;;
        --ocipush)
            OCIPUSH="$2"
            shift 2
            ;;
        --helmrepo-url)
            HELMREPO_URL="$2"
            shift 2
            ;;
        --kustomize)
            KUSTOMIZE=true
            shift
            ;;
        --gitrepo)
            GITREPO="$2"
            shift 2
            ;;
        --path)
            KUSTOMIZE_PATH="$2"
            shift 2
            ;;
        --ref)
            GITREF="$2"
            shift 2
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --skip-validate)
            SKIP_VALIDATE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 --appname <name> --version <version> [options]"
            echo ""
            echo "Modes:"
            echo "  Helm (OCI):    --ocirepo <oci-path>"
            echo "  Helm (repo):   --helmrepo <repo/chart> --ocipush <oci-path> [--helmrepo-url <url>]"
            echo "  Kustomize:     --kustomize --gitrepo <url> --path <path> [--ref <branch|tag>]"
            echo ""
            echo "Options:"
            echo "  --appname       Application name (e.g., podinfo, katib)"
            echo "  --version       Application version (e.g., 6.9.4, 0.17.0)"
            echo "  --ocirepo       OCI repository path - use when chart is already in OCI"
            echo "  --helmrepo      Helm repo/chart - use to pull from Helm repo and push to OCI"
            echo "  --ocipush       OCI path to push chart to (required with --helmrepo)"
            echo "  --helmrepo-url  Optional: Helm repo URL (required with --helmrepo if repo not added)"
            echo "  --kustomize     Use Kustomize-based install (GitRepository + Flux Kustomization)"
            echo "  --gitrepo       Git repo URL (required with --kustomize)"
            echo "  --path          Path within repo to kustomization.yaml (required with --kustomize)"
            echo "  --ref           Git ref: branch or tag (default: main)"
            echo "  --force         Skip confirmation prompt if application directory exists"
            echo "  --skip-validate Skip full catalog validation after add (run validate once after adding all apps)"
            echo ""
            echo "Examples:"
            echo "  $0 --appname podinfo --version 6.9.4 --ocirepo oci://ghcr.io/stefanprodan/charts/podinfo"
            echo "  $0 --appname katib --version 0.17.0 --kustomize --gitrepo https://github.com/kubeflow/manifests --path ./apps/katib/overlays/default --ref release-v1.10"
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$APPNAME" ]; then
    echo -e "${RED}Error: --appname is required${NC}"
    echo "Use --help for usage information"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: --version is required${NC}"
    echo "Use --help for usage information"
    exit 1
fi

# Mode validation: kustomize vs helm
if [ "$KUSTOMIZE" = true ]; then
    if [ -z "$GITREPO" ] || [ -z "$KUSTOMIZE_PATH" ]; then
        echo -e "${RED}Error: --gitrepo and --path are required when using --kustomize${NC}"
        exit 1
    fi
    if [ -n "$OCIREPO" ] || [ -n "$HELMREPO" ]; then
        echo -e "${RED}Error: --kustomize cannot be used with --ocirepo or --helmrepo${NC}"
        exit 1
    fi
else
    # Helm mode: must have either ocirepo OR (helmrepo + ocipush)
    if [ -n "$OCIREPO" ] && [ -n "$HELMREPO" ]; then
        echo -e "${RED}Error: Use either --ocirepo or --helmrepo, not both${NC}"
        exit 1
    fi

    if [ -n "$HELMREPO" ]; then
        if [ -z "$OCIPUSH" ]; then
            echo -e "${RED}Error: --ocipush is required when using --helmrepo${NC}"
            exit 1
        fi
        if [[ ! "$OCIPUSH" =~ ^oci:// ]]; then
            OCIPUSH="oci://${OCIPUSH}"
        fi
    elif [ -z "$OCIREPO" ]; then
        echo -e "${RED}Error: Use --ocirepo, (--helmrepo + --ocipush), or --kustomize with --gitrepo and --path${NC}"
        exit 1
    fi
fi

# If using helmrepo: pull, push, then set OCIREPO
if [ -n "$HELMREPO" ]; then
    echo -e "${BLUE}Pulling chart from Helm repo and pushing to OCI...${NC}"
    REPO_DIR_TMP="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PULL_DIR=$(mktemp -d)
    trap "rm -rf '$PULL_DIR'" EXIT

    # Add Helm repo if URL provided
    if [ -n "$HELMREPO_URL" ]; then
        REPO_NAME="${HELMREPO%%/*}"
        echo -e "${YELLOW}Adding Helm repo: ${REPO_NAME} -> ${HELMREPO_URL}${NC}"
        helm repo add "$REPO_NAME" "$HELMREPO_URL" 2>/dev/null || true
        helm repo update "$REPO_NAME"
    fi

    # Pull chart into temp dir
    echo -e "${YELLOW}Pulling ${HELMREPO}@${VERSION}...${NC}"
    (cd "$PULL_DIR" && helm pull "$HELMREPO" --version "$VERSION")
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to pull chart${NC}"
        exit 1
    fi

    CHART_TGZ=$(ls -1 "$PULL_DIR"/*.tgz 2>/dev/null | head -1)
    if [ -z "$CHART_TGZ" ] || [ ! -f "$CHART_TGZ" ]; then
        echo -e "${RED}Failed to find pulled chart .tgz file${NC}"
        exit 1
    fi

    # Get chart name from Chart.yaml (handles karmada-operator-chart-v1.16.0.tgz etc.)
    CHART_NAME=$(helm show chart "$CHART_TGZ" 2>/dev/null | grep '^name:' | awk '{print $2}')
    if [ -z "$CHART_NAME" ]; then
        # Fallback: derive from tgz filename (strip .tgz and version suffix)
        BASE=$(basename "$CHART_TGZ" .tgz)
        CHART_NAME="${BASE%-${VERSION}}"
        CHART_NAME="${CHART_NAME%-v${VERSION}}"
        [ -z "$CHART_NAME" ] && CHART_NAME="$BASE"
    fi

    # Push to OCI
    echo -e "${YELLOW}Pushing chart to ${OCIPUSH}...${NC}"
    helm push "$CHART_TGZ" "$OCIPUSH"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to push chart to OCI${NC}"
        exit 1
    fi

    # OCI path for add-application: ocipush/chartname
    OCIREPO="${OCIPUSH}/${CHART_NAME}"
    echo -e "${GREEN}✓ Chart pushed to ${OCIREPO}${NC}"
    echo ""
fi

# Validate OCI repo format for Helm mode
if [ "$KUSTOMIZE" = false ]; then
    if [ -z "$OCIREPO" ]; then
        echo -e "${RED}Error: OCIREPO not set${NC}"
        exit 1
    fi
    if [[ ! "$OCIREPO" =~ ^oci:// ]]; then
        echo -e "${YELLOW}Warning: OCI repo path should start with 'oci://'. Adding it automatically...${NC}"
        OCIREPO="oci://${OCIREPO}"
    fi
fi

# Get the repository root directory
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="${REPO_DIR}/applications/${APPNAME}/${VERSION}"
HELMR_DIR="${APP_DIR}/helmrelease"
METADATA_FILE="${APP_DIR}/metadata.yaml"

# Use local nkp binary if present, otherwise expect it in PATH
NKP_CMD="${REPO_DIR}/nkp"
if [ ! -f "$NKP_CMD" ] || [ ! -x "$NKP_CMD" ]; then
    NKP_CMD="nkp"
fi

# Load .env.local if it exists (for nkp validate / catalog operations)
if [ -f "${REPO_DIR}/.env.local" ]; then
    set -a
    source "${REPO_DIR}/.env.local"
    set +a
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Adding Application to NKP Catalog${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Application: ${GREEN}${APPNAME}${NC}"
echo -e "Version: ${GREEN}${VERSION}${NC}"
if [ "$KUSTOMIZE" = true ]; then
    echo -e "Mode: ${GREEN}Kustomize (GitRepository + Flux Kustomization)${NC}"
    echo -e "Git repo: ${GREEN}${GITREPO}${NC}"
    echo -e "Path: ${GREEN}${KUSTOMIZE_PATH}${NC}"
    echo -e "Ref: ${GREEN}${GITREF}${NC}"
else
    echo -e "OCI Repository: ${GREEN}${OCIREPO}${NC}"
fi
echo ""

# --- Kustomize mode ---
if [ "$KUSTOMIZE" = true ]; then
    echo -e "${YELLOW}Step 1: Creating Kustomize-based application structure...${NC}"
    if [ -d "$APP_DIR" ]; then
        echo -e "${YELLOW}Warning: Application directory already exists: ${APP_DIR}${NC}"
        if [ "$FORCE" = false ]; then
            read -p "Do you want to continue? This may overwrite existing files. (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${RED}Aborted by user${NC}"
                exit 1
            fi
        else
            echo -e "${YELLOW}  --force flag set, proceeding without confirmation${NC}"
        fi
    fi

    mkdir -p "$HELMR_DIR"

    # Root kustomization.yaml
    cat > "${APP_DIR}/kustomization.yaml" << 'ROOT_KUST_EOF'
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- helmrelease.yaml
ROOT_KUST_EOF

    # Root helmrelease.yaml (Flux Kustomization - same pattern as Helm apps)
    cat > "${APP_DIR}/helmrelease.yaml" << 'ROOT_HELM_EOF'
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: ${releaseName}-helmrelease
  namespace: ${releaseNamespace}
spec:
  interval: 6h0m0s
  path: ./helmrelease
  postBuild:
    substitute:
      releaseName: ${releaseName}
      releaseNamespace: ${releaseNamespace}
  prune: true
  retryInterval: 1m0s
  sourceRef:
    kind: OCIRepository
    name: ${releaseName}-source
    namespace: ${releaseNamespace}
  timeout: 10m0s
  wait: true
---
ROOT_HELM_EOF

    # GitRepository in helmrelease/
    # Use tag if ref looks like a version tag, else branch
    REF_TYPE="branch"
    [[ "$GITREF" =~ ^v[0-9] ]] || [[ "$GITREF" =~ ^release- ]] || [[ "$GITREF" =~ ^[0-9]+\.[0-9]+ ]] && REF_TYPE="tag"

    cat > "${HELMR_DIR}/gitrepository.yaml" << GITREPO_EOF
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: \${releaseName}-manifests
  namespace: \${releaseNamespace}
spec:
  interval: 1h
  url: ${GITREPO}
  ref:
    ${REF_TYPE}: ${GITREF}
---
GITREPO_EOF

    # Flux Kustomization in helmrelease/
    cat > "${HELMR_DIR}/flux-kustomization.yaml" << FLUX_KUST_EOF
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: \${releaseName}-kustomize
  namespace: \${releaseNamespace}
spec:
  interval: 10m
  path: ${KUSTOMIZE_PATH}
  prune: true
  sourceRef:
    kind: GitRepository
    name: \${releaseName}-manifests
    namespace: \${releaseNamespace}
  targetNamespace: \${releaseNamespace}
  timeout: 5m
  wait: true
---
FLUX_KUST_EOF

    # Kustomize kustomization.yaml in helmrelease/
    cat > "${HELMR_DIR}/kustomization.yaml" << 'HELMR_KUST_EOF'
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- gitrepository.yaml
- flux-kustomization.yaml
HELMR_KUST_EOF

    # metadata.yaml
    DISPLAY_NAME=$(echo "$APPNAME" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2)); print}')
    # Derive support link from git repo URL (e.g. https://github.com/kubeflow/manifests -> https://github.com/kubeflow)
    SUPPORT_LINK="${GITREPO%/manifests*}"
    [[ -z "$SUPPORT_LINK" || "$SUPPORT_LINK" == "$GITREPO" ]] && SUPPORT_LINK="https://github.com/${APPNAME}/${APPNAME}"

    cat > "$METADATA_FILE" << METADATA_EOF
schema: catalog.nkp.nutanix.com/v1/application-metadata
allowMultipleInstances: false
category:
- ai
- infrastructure
dependencies: []
description: |
  Kustomize-based deployment from ${GITREPO}.
  Path: ${KUSTOMIZE_PATH}
  Ref: ${GITREF}
  Update metadata.yaml with detailed description.
displayName: ${DISPLAY_NAME}
icon: ""
licensing:
- Pro
- Ultimate
overview: |
  Kustomize-based GitOps deployment. Manifests fetched from Git and applied via Flux Kustomization.
  Update metadata.yaml with detailed overview.
scope:
- workspace
- project
supportLink: ${SUPPORT_LINK}
METADATA_EOF

    echo -e "${GREEN}✓ Kustomize application structure created${NC}"
    echo ""

    # Skip nkp validate for Kustomize (it validates OCI charts; run ./validate.sh manually if needed)
    echo -e "${YELLOW}Step 2: Validation${NC}"
    echo -e "${GREEN}✓ Structure created. Run ./catalog-workflow.sh validate to validate the catalog.${NC}"
    echo ""

    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}Kustomize application added successfully!${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "Application directory: ${GREEN}${APP_DIR}${NC}"
    echo -e "GitRepository: ${GREEN}${HELMR_DIR}/gitrepository.yaml${NC}"
    echo -e "Flux Kustomization: ${GREEN}${HELMR_DIR}/flux-kustomization.yaml${NC}"
    echo -e "Metadata: ${GREEN}${METADATA_FILE}${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Review and update ${METADATA_FILE} with detailed description and supportLink"
    echo "  2. If ref should be a tag (not branch), edit ${HELMR_DIR}/gitrepository.yaml: change ref.branch to ref.tag"
    echo "  3. Add postBuild.substitute or patches in flux-kustomization.yaml if needed"
    echo "  4. Run validation: ./catalog-workflow.sh validate"
    echo "  5. Build and push: ./catalog-workflow.sh build-push <tag>"
    echo ""
    exit 0
fi

# --- Helm mode ---
HELMRELEASE_FILE="${APP_DIR}/helmrelease/helmrelease.yaml"

# Step 1: Generate catalog repository structure
echo -e "${YELLOW}Step 1: Generating catalog repository structure...${NC}"
if [ -d "$APP_DIR" ]; then
    echo -e "${YELLOW}Warning: Application directory already exists: ${APP_DIR}${NC}"
    if [ "$FORCE" = false ]; then
        read -p "Do you want to continue? This may overwrite existing files. (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Aborted by user${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}  --force flag set, proceeding without confirmation${NC}"
    fi
fi

"$NKP_CMD" generate catalog-repository --apps="${APPNAME}=${VERSION}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to generate catalog repository structure!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Catalog repository structure generated${NC}"
echo ""

# Step 2: Update OCI repository path in helmrelease.yaml
echo -e "${YELLOW}Step 2: Updating OCI repository path in helmrelease.yaml...${NC}"
if [ ! -f "$HELMRELEASE_FILE" ]; then
    echo -e "${RED}Error: helmrelease.yaml not found at ${HELMRELEASE_FILE}${NC}"
    exit 1
fi

# Use sed to replace the OCI repository URL
# This looks for the url: field in the OCIRepository resource and replaces it
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    sed -i '' "s|url: oci://.*|url: ${OCIREPO}|" "$HELMRELEASE_FILE"
else
    # Linux uses GNU sed
    sed -i "s|url: oci://.*|url: ${OCIREPO}|" "$HELMRELEASE_FILE"
fi

echo -e "${GREEN}✓ OCI repository path updated${NC}"
echo ""

# Step 3: Update metadata.yaml with basic information
echo -e "${YELLOW}Step 3: Updating metadata.yaml...${NC}"
if [ ! -f "$METADATA_FILE" ]; then
    echo -e "${RED}Error: metadata.yaml not found at ${METADATA_FILE}${NC}"
    exit 1
fi

# Create a capitalized display name from appname
# Convert kebab-case to Title Case
DISPLAY_NAME=$(echo "$APPNAME" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2)); print}')

# Update displayName
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    sed -i '' "s/^displayName:.*/displayName: ${DISPLAY_NAME}/" "$METADATA_FILE"
else
    # Linux uses GNU sed
    sed -i "s/^displayName:.*/displayName: ${DISPLAY_NAME}/" "$METADATA_FILE"
fi

# Check if supportLink is empty and set a default GitHub URL
CURRENT_SUPPORT_LINK=$(grep "^supportLink:" "$METADATA_FILE" | sed 's/supportLink: *//' | sed 's/"//g')
if [[ -z "$CURRENT_SUPPORT_LINK" ]] || [[ "$CURRENT_SUPPORT_LINK" == '""' ]] || [[ "$CURRENT_SUPPORT_LINK" == "''" ]]; then
    # Set a default GitHub URL based on appname (common pattern)
    DEFAULT_SUPPORT_LINK="https://github.com/${APPNAME}/${APPNAME}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|^supportLink:.*|supportLink: ${DEFAULT_SUPPORT_LINK}|" "$METADATA_FILE"
    else
        sed -i "s|^supportLink:.*|supportLink: ${DEFAULT_SUPPORT_LINK}|" "$METADATA_FILE"
    fi
    echo -e "${YELLOW}  Set default supportLink to: ${DEFAULT_SUPPORT_LINK}${NC}"
    echo -e "${YELLOW}  Note: You may want to update this with the correct support URL${NC}"
fi

# Note: The description field in metadata.yaml is typically a multi-line YAML field
# We'll leave it as-is since nkp generate should create a reasonable default
# Users can manually update it with more detailed information

echo -e "${GREEN}✓ Metadata.yaml updated with basic information${NC}"
echo -e "${YELLOW}  Note: You may want to manually update metadata.yaml with more detailed information${NC}"
echo ""

# Step 4: Validate the catalog repository (unless --skip-validate)
if [ "$SKIP_VALIDATE" = true ]; then
    echo -e "${YELLOW}Step 4: Skipping validation (--skip-validate). Run ./catalog-workflow.sh validate after adding all apps.${NC}"
else
    echo -e "${YELLOW}Step 4: Validating catalog repository...${NC}"
    "$NKP_CMD" validate catalog-repository --repo-dir="${REPO_DIR}"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Validation failed!${NC}"
        echo -e "${YELLOW}Please check the errors above and fix them manually${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Validation passed${NC}"
fi
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Application added successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Application directory: ${GREEN}${APP_DIR}${NC}"
echo -e "HelmRelease file: ${GREEN}${HELMRELEASE_FILE}${NC}"
echo -e "Metadata file: ${GREEN}${METADATA_FILE}${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review and update ${METADATA_FILE} with detailed information"
echo "  2. Review ${HELMRELEASE_FILE} to ensure OCI repository path is correct"
echo "  3. Update any configuration in ${APP_DIR}/helmrelease/cm.yaml if needed"
echo "  4. Run validation again: nkp validate catalog-repository --repo-dir=."
echo "  5. Build and push the catalog: ./catalog-workflow.sh build-push --tag <tag>"
echo ""

