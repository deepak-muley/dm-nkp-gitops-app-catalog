#!/bin/sh
# Usage: ./add-catalog-to-cluster.sh [workspace-name]
# Example: ./add-catalog-to-cluster.sh dm-dev-workspace

WORKSPACE="${1:-dm-dev-workspace}"

nkp create catalog-collection \
  --url oci://ghcr.io/deepak-muley/nkp-custom-apps-catalog/dm-nkp-gitops-app-catalog/collection \
  --tag v0.1.0 \
  --workspace "$WORKSPACE"
