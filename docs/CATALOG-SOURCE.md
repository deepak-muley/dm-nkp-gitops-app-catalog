# Catalog source metadata (Helm-repo → OCI apps)

For apps that you add with **Helm repo + OCI push** (`--helmrepo` and `--ocipush`), new chart versions are published in the **original Helm repo**, not in your OCI registry. The `check-versions` command needs to know where to look.

## Per-app file: `.catalog-source.yaml`

Place a file **`applications/<app>/.catalog-source.yaml`** (in the app folder, not inside a version folder) with:

```yaml
# Chart is pulled from this Helm repo, then pushed to OCI. check-versions uses this to find latest version.
helmrepo: kyverno/kyverno
helmrepoUrl: https://kyverno.github.io/kyverno/
ocipush: oci://ghcr.io/YOUR_ORG/kyverno
```

| Field         | Required | Description |
|---------------|----------|-------------|
| `helmrepo`    | Yes      | `repo_name/chart_name` as used in `helm search repo` and `helm pull`. |
| `helmrepoUrl` | Yes      | Helm repo URL (e.g. `https://kyverno.github.io/kyverno/`). Used by `check-versions` to run `helm repo add` / `helm search repo`. |
| `ocipush`     | Yes      | OCI base path you push to (e.g. `oci://ghcr.io/YOUR_ORG/kyverno`). Used to print the exact `add-app` command. |

## When to add it

- You use **`add-app --helmrepo X --ocipush Y [--helmrepo-url Z]`** for this app.
- You want **`./catalog-workflow.sh check-versions --all`** (or `--appname <app>`) to check the **upstream Helm repo** for new versions and recommend the add-app command.

If you do **not** add this file, `check-versions` will treat the app as **OCI-only**: it will use the OCI URL from `helmrelease/helmrelease.yaml` and list tags from that registry. That is correct when the chart is published directly to OCI (e.g. `oci://ghcr.io/grafana/helm-charts/loki`).

## Example apps using `.catalog-source.yaml`

- `applications/kyverno/.catalog-source.yaml`
- `applications/kubescape-operator/.catalog-source.yaml`
- `applications/vault/.catalog-source.yaml`
- `applications/karmada-operator/.catalog-source.yaml`

## Adding a new Helm-repo app

1. Add the app once with `add-app` as usual:
   ```bash
   ./catalog-workflow.sh add-app --appname my-chart --version 1.0.0 \
     --helmrepo myrepo/my-chart --ocipush oci://ghcr.io/YOUR_ORG/my-chart \
     --helmrepo-url https://myrepo.github.io/helm-charts/
   ```
2. Create `applications/my-chart/.catalog-source.yaml` with the same `helmrepo`, `helmrepoUrl`, and `ocipush`.
3. After that, `check-versions --appname my-chart` (or `--all`) will check the Helm repo for newer versions and suggest the add-app command.
