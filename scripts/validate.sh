#!/bin/bash

# Script to validate NKP catalog repository (with ghcr.io login for private OCI charts)
# Usage: ./validate.sh

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Load credentials from .env.local
if [ -f "${REPO_DIR}/.env.local" ]; then
    source "${REPO_DIR}/.env.local"
fi

# Login to ghcr.io so nkp validate can pull private OCI charts
if [ -n "$GHCR_USERNAME" ] && [ -n "$GHCR_PASSWORD" ]; then
    echo "Logging into ghcr.io..."
    echo "${GHCR_PASSWORD}" | docker login ghcr.io -u "${GHCR_USERNAME}" --password-stdin
fi

# Use local nkp binary if present, otherwise expect it in PATH
NKP_CMD="${REPO_DIR}/nkp"
if [ ! -f "$NKP_CMD" ] || [ ! -x "$NKP_CMD" ]; then
    NKP_CMD="nkp"
fi

"$NKP_CMD" validate catalog-repository --repo-dir="${REPO_DIR}"
