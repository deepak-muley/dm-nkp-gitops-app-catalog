#!/bin/sh
# Push an existing catalog bundle using the NKP CLI.
# Usage: ./push-bundle.sh <tag>
# Requires: nkp in PATH (with push bundle), GHCR_USERNAME and GHCR_PASSWORD in env or .env.local.

set -e
REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$REPO_DIR"

[ -f .env.local ] && set -a && . ./.env.local && set +a

TAG="${1:?Usage: ./push-bundle.sh <tag>   e.g. ./push-bundle.sh v0.7.0}"
REGISTRY="oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog"

if [ -f "${REPO_DIR}/dm-nkp-gitops-app-catalog-${TAG}.tar" ]; then
    BUNDLE_FILE="${REPO_DIR}/dm-nkp-gitops-app-catalog-${TAG}.tar"
elif [ -f "${REPO_DIR}/dm-nkp-gitops-app-catalog.tar" ]; then
    BUNDLE_FILE="${REPO_DIR}/dm-nkp-gitops-app-catalog.tar"
else
    echo "Bundle not found: dm-nkp-gitops-app-catalog-${TAG}.tar or dm-nkp-gitops-app-catalog.tar"
    exit 1
fi

echo "${GHCR_PASSWORD}" | docker login ghcr.io -u "${GHCR_USERNAME}" --password-stdin
nkp push bundle "${BUNDLE_FILE}" --to-registry "${REGISTRY}" \
    --to-registry-username "${GHCR_USERNAME}" --to-registry-password "${GHCR_PASSWORD}"
