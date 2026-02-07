# Claude / AI Agent Instructions — dm-nkp-gitops-app-catalog

Instructions for Claude Code and similar AI assistants. **Assumption:** This file and `AGENTS.md` are kept up to date; agents can rely on them as the source of truth.

## Project

NKP custom app catalog for Nutanix Kubernetes Platform. Applications are Flux/Kustomize manifests with catalog metadata.

## Essential Commands

- **Add application**: `./add-application.sh --appname <name> --version <version> --ocirepo oci://<registry>/path/chart`
- **Validate**: `./validate.sh`
- **Build & push**: `./build-and-push.sh <tag>`
- **Credentials**: `source .env.local` (or scripts load it automatically)

## Rules for AI Assistants

1. **Use catalog-workflow** — Always prefer `./catalog-workflow.sh add-app` when adding apps; use `validate` for validation.
2. **No manual app creation** for standard OCI Helm chart apps — use `./catalog-workflow.sh add-app`.
3. **Job-based apps** (e.g. kagent): No OCI Helm chart exists. Copy structure from `applications/kagent/0.1.0/` and adapt.
4. **Do not check in** `.env.local`, `setup-credentials.sh`, or any secrets.
5. **Chart versions** — Ensure OCI chart tag exists; validation will fail on invalid tags.
6. **metadata.yaml** — Schema `catalog.nkp.nutanix.com/v1/application-metadata`; set `supportLink` to project URL (e.g. https://kgateway.dev).

## Validation Notes

- `nkp validate` pulls OCI artifacts; requires `docker login ghcr.io` for private charts.
- `./catalog-workflow.sh validate` handles login and uses local `./nkp` when present.

## Reference

See `AGENTS.md` for a concise, agent-agnostic version of these instructions.
