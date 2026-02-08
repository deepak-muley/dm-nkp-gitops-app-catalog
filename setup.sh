#!/usr/bin/env bash
# Setup apptests — ensure Go dependencies are available for this catalog's tests.
# Usage: ./setup.sh
# Requires: go, docker (for Kind)
# Tests are specific to this catalog (applications/ layout). No external clone.
# See apptests/README.md and docs/APP-TESTS-GUIDE.md

set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APPTESTS_DIR="${REPO_DIR}/apptests"

echo "=== Apptests Setup ==="
echo ""

if [[ ! -f "$APPTESTS_DIR/go.mod" ]]; then
  echo "Error: apptests/go.mod not found. Apptests structure should be in apptests/"
  echo "See docs/APP-TESTS-GUIDE.md for creating apptests."
  exit 1
fi

echo "Running go mod tidy..."
(cd "$APPTESTS_DIR" && go mod tidy)
if [[ $? -ne 0 ]]; then
  echo "Error: go mod tidy failed"
  exit 1
fi

echo ""
echo "=== Setup complete ==="
echo ""
echo "Run apptests:"
echo "  just apptests              # all tests"
echo "  just apptests-app podinfo  # podinfo only"
echo ""
echo "Without just: cd apptests && go test ./suites/ -v -run podinfo"
echo ""
echo "See apptests/README.md and docs/APP-TESTS-GUIDE.md for details."
