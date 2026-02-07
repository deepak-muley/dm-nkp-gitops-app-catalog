# Agent Instructions — dm-nkp-gitops-app-catalog

This file provides context for AI coding agents (Cursor, Claude, etc.) working in this repository. **Assumption:** This file and `CLAUDE.md` are kept up to date; agents can rely on them as the source of truth for project commands and conventions.

## Project

NKP (Nutanix Kubernetes Platform) custom app catalog. Applications are Flux/Kustomize-based with metadata for the NKP catalog UI.

## Commands

| Action | Command |
|--------|---------|
| **Workflow (all-in-one)** | `./catalog-workflow.sh <add-app\|validate\|add-tests\|build-push\|all>` |
| Add app (OCI Helm chart) | `./catalog-workflow.sh add-app --appname X --version Y --ocirepo oci://...` |
| Validate catalog | `./catalog-workflow.sh validate` |
| Add test placeholders | `./catalog-workflow.sh add-tests --appname X` or `--all` |
| Build and push bundle | `./catalog-workflow.sh build-push --tag v0.1.0` |
| Validate + build-push | `./catalog-workflow.sh all --tag v0.1.0` |
| Load credentials | `source .env.local` |
| Run apptests (just) | `just apptests` \| `just apptests-app <name>` |

## Conventions

1. **Use existing scripts** — Prefer `add-application.sh`, `validate.sh`, `build-and-push.sh` over manual steps.
2. **Do not check in secrets** — `.env.local`, `setup-credentials.sh` are gitignored.
3. **Chart version** — OCI chart `ref.tag` must exist; validation fails otherwise.
4. **Two app patterns**:
   - **Helm-based** (default): OCIRepository + HelmRelease. Use `./catalog-workflow.sh add-app`.
   - **Job-based** (kagent): Flux Kustomization + Job. Copy from `applications/kagent/0.1.0/`.

## Structure

```
applications/<app>/<version>/
├── metadata.yaml       # catalog.nkp.nutanix.com/v1/application-metadata
├── helmrelease.yaml
├── kustomization.yaml
└── helmrelease/        # Flux resources
```

## Validation

- `./validate.sh` does Docker login + `nkp validate catalog-repository`.
- CI (`.github/workflows/ci.yml`) checks structure, YAML syntax, metadata schema.
- NKP validate pulls OCI charts; private repos need `docker login ghcr.io` first.

## nkp CLI

- Local binary at repo root: `./nkp` (or `nkp` in PATH).
- Used by catalog-workflow (scripts/add-application.sh, validate.sh, build-and-push.sh).
