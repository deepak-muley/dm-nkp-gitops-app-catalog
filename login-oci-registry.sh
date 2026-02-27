#!/bin/sh
set -a
. ./.env.local
set +a
export GHCR_USERNAME GHCR_PASSWORD

case "$0" in *login-oci-registry.sh)
  echo "Logging into ghcr.io..."
  docker login ghcr.io -u "$GHCR_USERNAME" -p "$GHCR_PASSWORD"
  echo "Done. To use GHCR_USERNAME/GHCR_PASSWORD in this shell, run: source login-oci-registry.sh"
  ;;
esac
