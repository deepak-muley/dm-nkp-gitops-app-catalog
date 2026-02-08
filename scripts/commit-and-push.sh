#!/usr/bin/env bash
# Run this script from your system terminal (outside Cursor) to commit and push
# when Cursor's git wrapper causes "unknown option trailer" with older Git.
set -e
cd "$(git rev-parse --show-toplevel)"
git add -A
if git diff --cached --quiet; then
  echo "Nothing to commit."
  exit 0
fi
git commit -m "Catalog metadata and rules: Table 41 schema, overview format, dm-nkp-gitops-app-catalog category, ClickStack and Spark Operator"
git push
