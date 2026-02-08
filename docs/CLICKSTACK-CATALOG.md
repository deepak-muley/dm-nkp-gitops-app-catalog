# ClickStack in NKP Catalog

This document describes ClickStack (the ClickHouse observability stack), how it is deployed via this catalog, and how to use it.

## What is ClickStack?

**ClickStack** is a production-grade observability platform built on [ClickHouse](https://clickhouse.com), unifying **logs**, **traces**, **metrics**, and **sessions** in a single high-performance solution. It is designed for monitoring and debugging complex systems and enables developers and SREs to trace issues end-to-end without switching between tools or manually stitching data together.

Reference: [ClickStack overview](https://clickhouse.com/docs/use-cases/observability/clickstack/overview).

### Core idea

All observability data is ingested as wide, rich events. These events are stored in ClickHouse tables by data type (logs, traces, metrics, sessions) but remain fully queryable and cross-correlatable at the database level. ClickStack leverages ClickHouse’s column-oriented architecture, native JSON support, and parallelized execution for sub-second queries and high-cardinality workloads.

### Components

| Component | Role |
|-----------|------|
| **ClickStack UI (HyperDX)** | Purpose-built frontend for exploring and visualizing observability data (log search, trace exploration, dashboards, alerting). |
| **OpenTelemetry collector** | Preconfigured collector with an opinionated schema for logs, traces, and metrics; writes to ClickHouse. |
| **ClickHouse** | High-performance analytical database at the heart of the stack. |
| **MongoDB** | Used by the default Helm chart for persistent application state (e.g. dashboards, user accounts). |

You can deploy the full stack (default) or integrate with an existing ClickHouse or ClickHouse Cloud instance.

---

## Use cases

- **Unified observability** — Correlate and search logs, metrics, session replays, and traces in one place; no need to jump between tools.
- **Root cause and debugging** — Trace from front-end sessions to backend infrastructure, application logs, and distributed traces.
- **Full-text and property search** — Use Lucene-style syntax (e.g. `level:err`) or SQL; dashboards and live tail for real-time inspection.
- **Alerting and anomaly detection** — Set up alerts and use event deltas to spot anomalies and performance regressions.
- **OpenTelemetry-native** — Ingest via OTLP; use the ClickStack distribution of the OTel collector or your own.
- **Cost-effective at scale** — ClickHouse’s compression and columnar storage support long retention and high cardinality without blowing cost.

---

## Adding ClickStack to this catalog

This catalog uses the [ClickHouse ClickStack Helm chart](https://github.com/ClickHouse/ClickStack-helm-charts). The chart is pulled from the Helm repo and pushed to OCI when adding the app.

### add-app command

```bash
./catalog-workflow.sh add-app --appname clickstack --version 1.1.1 \
  --helmrepo clickstack/clickstack \
  --helmrepo-url https://clickhouse.github.io/ClickStack-helm-charts \
  --ocipush oci://ghcr.io/deepak-muley/charts/clickstack \
  --force
```

You need `docker login ghcr.io` (or your OCI registry) with push access for the `--ocipush` path.

### Check for new versions

```bash
./catalog-workflow.sh check-versions --appname clickstack
```

---

## Deployment via NKP catalog

When you deploy ClickStack from the NKP catalog UI or via Flux/Kommander:

1. Choose the **ClickStack** app and version **1.1.1**.
2. Set **release name** and **namespace** (e.g. `clickstack`, `observability`).
3. Optionally override Helm values (see below).

By default the chart installs ClickHouse, HyperDX (UI), the OpenTelemetry collector, and MongoDB. For production, consider disabling the in-chart ClickHouse or OTel collector and using your own or ClickHouse Cloud.

### Accessing the UI (HyperDX)

After install, the HyperDX UI is typically exposed by the chart (e.g. Service on port 3000). To access it:

- **Port-forward (dev / one-off):**
  ```bash
  kubectl port-forward -n <releaseNamespace> \
    pod/$(kubectl get pod -n <releaseNamespace> -l app.kubernetes.io/name=clickstack -o jsonpath='{.items[0].metadata.name}') \
    8080:3000
  ```
  Then open http://localhost:8080 and create a user; the UI will create data sources for the in-chart ClickHouse.

- **Production:** Configure ingress with TLS (see [Helm configuration](https://clickhouse.com/docs/use-cases/observability/clickstack/deployment/helm-configuration#ingress-setup)).

A **dashboard ConfigMap** (`clickstack-ui-dashboard-cm.yaml`) is included so the NKP/Kommander UI can show a link to the ClickStack (HyperDX) dashboard when the app is installed.

### Key Helm values (examples)

| Area | Example values |
|------|----------------|
| **Use ClickHouse Cloud** | `clickhouse.enabled: false`, set `otel.clickhouseEndpoint` and credentials. |
| **Production (external ClickHouse/OTel)** | `clickhouse.enabled: false`, `otel.enabled: false`; configure HyperDX to use your ClickHouse. |
| **Resources** | `replicaCount`, `resources.limits/requests` per component. |
| **Ingress** | `ingress.enabled: true`, `ingress.hosts`, TLS annotations. |
| **Secrets** | Use existing secrets for API keys / DB credentials via `hyperdx.apiKey.valueFrom.secretKeyRef`, etc. |

See the [Helm configuration guide](https://clickhouse.com/docs/use-cases/observability/clickstack/deployment/helm-configuration) and the chart [values](https://github.com/ClickHouse/ClickStack-helm-charts) for full options.

---

## Sending data (OpenTelemetry)

- Point your applications and infrastructure to the OpenTelemetry collector endpoint (OTLP) deployed by the chart, or run your own OTel collector and configure it to export to ClickHouse.
- The [ClickStack distribution](https://clickhouse.com/docs/use-cases/observability/clickstack/overview) of the collector is preconfigured for ClickHouse ingestion.

---

## Documentation and references

- [ClickStack overview](https://clickhouse.com/docs/use-cases/observability/clickstack/overview)
- [Getting started (Helm)](https://clickhouse.com/docs/use-cases/observability/clickstack/deployment/helm)
- [Helm configuration](https://clickhouse.com/docs/use-cases/observability/clickstack/deployment/helm-configuration)
- [ClickStack Helm charts (GitHub)](https://github.com/ClickHouse/ClickStack-helm-charts)
