#!/bin/bash

# Script to build and push NKP catalog bundle
# Usage: ./build-and-push.sh <tag>
# Example: ./build-and-push.sh v0.1.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if tag is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Tag is required${NC}"
    echo "Usage: $0 <tag>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

TAG="$1"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUNDLE_FILE="${REPO_DIR}/dm-nkp-gitops-app-catalog.tar"
REGISTRY="oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog"

# Load .env.local if it exists
if [ -f "${REPO_DIR}/.env.local" ]; then
    echo -e "${YELLOW}Loading environment from .env.local...${NC}"
    source "${REPO_DIR}/.env.local"
fi

# Check for required environment variables
if [ -z "$GHCR_USERNAME" ]; then
    echo -e "${RED}Error: GHCR_USERNAME environment variable is not set${NC}"
    echo "Please set it: export GHCR_USERNAME=deepak-muley"
    exit 1
fi

if [ -z "$GHCR_PASSWORD" ]; then
    echo -e "${RED}Error: GHCR_PASSWORD environment variable is not set${NC}"
    echo "Please set it: export GHCR_PASSWORD=<your-github-pat>"
    exit 1
fi

echo -e "${GREEN}Building and pushing catalog bundle with tag: ${TAG}${NC}"
echo ""

# Step 1: Validate catalog repository
echo -e "${YELLOW}Step 1: Validating catalog repository...${NC}"
nkp validate catalog-repository --repo-dir="${REPO_DIR}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Validation failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Validation passed${NC}"
echo ""

# Step 2: Create catalog bundle
echo -e "${YELLOW}Step 2: Creating catalog bundle...${NC}"
nkp create catalog-bundle --collection-tag "${TAG}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Bundle creation failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Bundle created: ${BUNDLE_FILE}${NC}"
echo ""

# Step 3: Login to GHCR
echo -e "${YELLOW}Step 3: Logging into GitHub Container Registry...${NC}"
echo "${GHCR_PASSWORD}" | docker login ghcr.io -u "${GHCR_USERNAME}" --password-stdin
if [ $? -ne 0 ]; then
    echo -e "${RED}Docker login to ghcr.io failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Logged into ghcr.io${NC}"
echo ""

# Step 4: Push bundle to registry
echo -e "${YELLOW}Step 4: Pushing bundle to registry...${NC}"
nkp push bundle \
    --bundle "${BUNDLE_FILE}" \
    --to-registry "${REGISTRY}" \
    --to-registry-username "${GHCR_USERNAME}" \
    --to-registry-password "${GHCR_PASSWORD}"
if [ $? -ne 0 ]; then
    echo -e "${RED}Push failed!${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Bundle pushed successfully${NC}"
echo ""

# Step 5: Make package public (optional)
if [ "$MAKE_PUBLIC" = "true" ]; then
    echo -e "${YELLOW}Step 5: Making packages public...${NC}"
    PACKAGE_TYPE="container"

    # List of packages to make public (URL-encoded names)
    PACKAGES=(
        "nkp-custom-apps-catalog%2Fdm-nkp-gitops-app-catalog%2Fcollection"
        "nkp-custom-apps-catalog%2Fdm-nkp-gitops-app-catalog%2Fkubescape-operator"
        "nkp-custom-apps-catalog%2Fdm-nkp-gitops-app-catalog%2Fkyverno"
        "nkp-custom-apps-catalog%2Fdm-nkp-gitops-app-catalog%2Fpodinfo"
        "nkp-custom-apps-catalog%2Fdm-nkp-gitops-app-catalog%2Fvault"
    )

    # Make each package public
    for PACKAGE_NAME in "${PACKAGES[@]}"; do
        DISPLAY_NAME=$(echo "$PACKAGE_NAME" | sed 's/%2F/\//g')

        # Check current visibility first
        CURRENT_VISIBILITY=$(curl -s \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${GHCR_PASSWORD}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            "https://api.github.com/user/packages/${PACKAGE_TYPE}/${PACKAGE_NAME}" \
            | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('visibility', 'unknown'))" 2>/dev/null)

        if [ "$CURRENT_VISIBILITY" = "public" ]; then
            echo -e "  ${GREEN}✓ ${DISPLAY_NAME} is already public${NC}"
            continue
        fi

        echo -e "  Making ${DISPLAY_NAME} public..."

        # Try PATCH endpoint
        RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${GHCR_PASSWORD}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            "https://api.github.com/user/packages/${PACKAGE_TYPE}/${PACKAGE_NAME}" \
            -d '{"visibility":"public"}')

        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
            echo -e "  ${GREEN}✓ ${DISPLAY_NAME} made public${NC}"
        else
            RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')
            echo -e "  ${YELLOW}⚠ Could not make ${DISPLAY_NAME} public (HTTP ${HTTP_CODE})${NC}"
            echo -e "    ${YELLOW}Note: You may need to make it public manually via GitHub UI${NC}"
            if [ -n "$RESPONSE_BODY" ] && [ "$RESPONSE_BODY" != "null" ]; then
                echo "    Response: $RESPONSE_BODY"
            fi
        fi
    done
    echo ""
fi

echo -e "${GREEN}✓ All steps completed successfully!${NC}"
echo ""
echo "Bundle URL: ${REGISTRY}/dm-nkp-gitops-app-catalog/collection:${TAG}"
echo ""
echo "To create catalog collection in NKP, run:"
echo "  nkp create catalog-collection --url ${REGISTRY}/dm-nkp-gitops-app-catalog/collection --tag ${TAG} --workspace <workspace-name>"

