# Application Metadata Fields

This document describes each field in the **NKP Application Metadata** schema (`catalog.nkp.nutanix.com/v1/application-metadata`) used in `applications/<app>/<version>/metadata.yaml`. These fields control how applications appear and behave in the NKP catalog UI.

---

## Required and identity

| Field | Required | Description |
|-------|----------|-------------|
| **schema** | Yes | Must be `catalog.nkp.nutanix.com/v1/application-metadata`. Identifies the metadata format for NKP. |

---

## Display and naming

| Field | Default | Description |
|-------|---------|-------------|
| **displayName** | (derived from app ID) | Human-readable name shown in the catalog UI (cards, lists, detail header). Use proper casing and branding (e.g. "Cert Manager", "vLLM"). |

---

## Instance and scope

| Field | Default | Description |
|-------|---------|-------------|
| **allowMultipleInstances** | `true` | If `true`, users can install more than one instance of the app (e.g. per namespace or environment). If `false`, only one instance is allowed (typical for cluster-wide operators). |
| **scope** | `[project]` | Where the app can be deployed. Use one or both of: `project`, `workspace`. Affects where the app is offered in the NKP UI. |

### allowMultipleInstances — when true vs false

| Value | Use when | Examples from this catalog |
|-------|----------|----------------------------|
| **`false`** | The app is a **cluster-wide singleton**: one instance per cluster (or per workspace/project scope). Multiple copies would conflict or are not meaningful. | **Cert Manager** (one issuer hierarchy per cluster), **Traefik** (one ingress controller), **Kyverno** (one policy engine), **Karmada Operator**, **Centralized OpenCost** (one backend per management cluster), **Spark Operator**, **Katib** |
| **`true`** | The app can be **safely installed multiple times**: per project, per namespace, or per team. Each instance is independent. | **Podinfo** (demo app per env), **OpenCost** (one per workload cluster with its own UI), **vLLM**, **KServe** (multiple model servers), **Loki**, **Kube Prometheus Stack**, **Vertical Pod Autoscaler** (per namespace possible) |

### scope — project vs workspace

| Value | Meaning | Examples from this catalog |
|-------|---------|----------------------------|
| **`project`** | App is offered when deploying at **project** level (e.g. a single cluster or project-scoped target). Typical for per-cluster or per-team tools. | **Podinfo**, **vLLM**, **KServe**, **Vertical Pod Autoscaler** — often `scope: [project]` only |
| **`workspace`** | App is offered when deploying at **workspace** level (e.g. management cluster or workspace-wide tooling). Typical for platform-level or multi-cluster services. | **Centralized OpenCost** (workspace only — backend on management cluster), **Traefik**, **Kyverno**, **Kubescape**, **Vault**, **Let's Encrypt ClusterIssuer** |
| **Both** `project` and `workspace` | App can be deployed at either level; choose based on where the UI shows it. | **Cert Manager**, **OpenCost**, **Kube Prometheus Stack**, **Loki**, **Katib**, **Kagent**, **ClickStack** |

### Relationship between allowMultipleInstances and scope

- **No strict coupling**: An app can be `allowMultipleInstances: false` with either `scope: workspace`, `scope: project`, or both. Same for `true`.
- **Typical patterns**:
  - **Singleton at workspace**: `allowMultipleInstances: false`, `scope: [workspace]` — e.g. one Centralized OpenCost per management cluster, one Traefik per workspace.
  - **Singleton at project**: `allowMultipleInstances: false`, `scope: [project]` — one cert-manager or Spark Operator per project/cluster.
  - **Multi-instance at project**: `allowMultipleInstances: true`, `scope: [project]` — e.g. multiple OpenCost or vLLM deployments per project.
  - **Available in both scopes**: Use both `workspace` and `project` in `scope` when the app is valid at either level; instance count is still controlled separately by `allowMultipleInstances`.

---

## Categorization and discovery

| Field | Default | Description |
|-------|---------|-------------|
| **category** | `[general]` | List of category slugs used to group and filter apps in the UI (e.g. `monitoring`, `ai-ml`, `security`). **This catalog requires** including `dm-nkp-gitops-app-catalog` so apps from this repo can be filtered. Add other categories as appropriate (e.g. `observability`, `cost`). |

---

## Descriptions and documentation

| Field | Default | Description |
|-------|---------|-------------|
| **description** | `""` | Short summary (one or two sentences) shown on the application card and in list views. Plain text or simple markdown; keep it concise. |
| **overview** | — | **Markdown** content shown on the application detail page. Use a consistent structure: "What it is", "Highlights", and links to docs/product. See [Overview format](#overview-format) below. |
| **supportLink** | — | URL to official documentation or support (e.g. project docs, GitHub). Used by the UI for "Support" or "Documentation" links. Prefer the primary project/docs URL. |

---

## Dependencies

| Field | Default | Description |
|-------|---------|-------------|
| **dependencies** | `[]` | List of application names that are recommended for this app to function. The UI may show these as suggestions; it does not block installation if they are missing. |
| **requiredDependencies** | — | List of application names that must be installed for this app to be **enabled** in the UI. If any are missing, the app may be disabled or hidden until dependencies are satisfied. |

---

## Compatibility and upgrades

| Field | Default | Description |
|-------|---------|-------------|
| **k8sVersionSupport** | — | String describing supported Kubernetes versions (e.g. `"1.29 to 1.32"`). Shown in the UI to help users check compatibility. |
| **nkpVersionSupport** | — | String describing supported NKP (Nutanix Kubernetes Platform) versions. Used for platform compatibility guidance. |
| **upgradesFrom** | — | Version or range of versions that this release can be upgraded from (e.g. for in-place upgrade flows). |

---

## Licensing and certifications

| Field | Default | Description |
|-------|---------|-------------|
| **licensing** | — | List of NKP license tiers required for the app to be installable. Common values: `Pro`, `Ultimate`. The cluster must have at least one of these licenses. |
| **certifications** | `[]` | List of certification labels (e.g. `qualified`) that NKP may use to badge or filter apps. |

---

## Visual and type

| Field | Default | Description |
|-------|---------|-------------|
| **icon** | — | URL to an image for the application logo (PNG, SVG, etc.), or a base64-encoded SVG. Shown on cards and in the detail view. Use `""` if no icon is available. |
| **type** | — | Application type. In this catalog, use `custom` for apps from this custom catalog. |

---

## Overview format

Keep the **overview** field consistent across apps. Suggested structure (Markdown):

1. **What it is** — One or two sentences describing the app and its purpose.
2. **Highlights** — Bullet list of key features or capabilities.
3. **Documentation / Product** — Links to official docs, product page, or GitHub.

Example:

```yaml
overview: |
  **What it is** — OpenCost is an open-source cost monitoring and allocation tool for Kubernetes.

  **Highlights**
  - Real-time cost allocation by namespace, deployment, and label
  - Integration with Prometheus
  - Optional UI for dashboards

  **Documentation:** [OpenCost docs](https://opencost.io/docs/) | **Project:** [GitHub](https://github.com/opencost/opencost)
```

---

## Validation

- Validate the catalog (including metadata) with:
  ```bash
  ./catalog-workflow.sh validate
  ```
- Or directly: `nkp validate catalog-repository --repo-dir=.`
- Required in practice for catalog quality: **displayName**, **category** (including `dm-nkp-gitops-app-catalog`), **description**, **supportLink**.

---

## Reference

- Schema identifier: `catalog.nkp.nutanix.com/v1/application-metadata`
- NKP platform documentation (e.g. Application Metadata Schema) for authoritative field semantics and UI behavior.
- In-repo: `.cursor/rules/application-structure.mdc` for structure and file layout; `.cursor/rules/catalog-category.mdc` for the `dm-nkp-gitops-app-catalog` category requirement.
