# claud.md — Agent Instructions

Alias/short reference for AI agents. For full instructions see `AGENTS.md` and `CLAUDE.md`.

## Quick Reference

| Task | Command |
|------|---------|
| Add app | `./catalog-workflow.sh add-app --appname X --version Y --ocirepo oci://...` |
| Validate | `./catalog-workflow.sh validate` |
| Build & push | `./catalog-workflow.sh build-push --tag <tag>` |

## Rules

1. Use existing scripts; do not manually create app dirs for Helm chart apps.
2. Never check in secrets (`.env.local`, etc.).
3. Ensure OCI chart tag exists before validation.
