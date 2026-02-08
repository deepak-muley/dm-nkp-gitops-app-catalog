# OpenCost in mgmt + workload setup

## What this catalog deploys

- **Management cluster**: `centralized-opencost` — same OpenCost chart with **UI disabled**. Used for cost visibility on the management cluster only (backend/API only).
- **Workload clusters**: `opencost` — OpenCost with **UI enabled**. Each workload cluster has its own cost data and UI.

Both use the same chart (`oci://ghcr.io/opencost/charts/opencost`). The only difference in our config is `opencost.ui.enabled` (false for central, true for workload). Each deployment uses **in-cluster Prometheus** (`prometheus.internal.enabled: true`).

## Does workload OpenCost connect to central OpenCost?

**No.** In this catalog there is **no** connection configured between workload OpenCost and centralized OpenCost. They do not talk to each other.

- Each cluster runs its own OpenCost and its own (internal) Prometheus.
- Workload clusters do not push cost data to the management cluster.
- The name “centralized-opencost” means “the OpenCost deployment on the management cluster (no UI),” not “the aggregator that workload OpenCost sends data to.”

So today: **mgmt** shows cost for the mgmt cluster only; **workload1** and **workload2** each show cost for their own cluster only.

## How you could get real multi-cluster aggregation

Aggregation is done at the **metrics** layer, not by OpenCost talking to OpenCost:

1. **Central Prometheus/Thanos**
   - Each cluster’s Prometheus remote-writes to a central Thanos (or central Prometheus) with a distinct `cluster` (or `cluster_id`) label.
   - That gives one place that has metrics from all clusters.

2. **OpenCost and central metrics**
   - OpenCost (community) is designed to query **one** Prometheus. It does not natively support querying a central store and splitting by cluster in a single deployment.
   - Common patterns:
     - **One OpenCost per cluster** (what we do): each cluster has its own cost view; no single “global” dashboard.
     - **Central Prometheus/Thanos + one OpenCost**: run one OpenCost instance (e.g. on mgmt) that points at the central Prometheus/Thanos; you need metrics to be labeled by cluster and, depending on version, possible limitations on filtering by cluster in the open-source version.
   - Enterprise offerings (e.g. Kubecost) add federated ETL (e.g. S3 + aggregator) for multi-cluster cost aggregation.

So if you want “one place to see all clusters’ cost,” you need:

- A central metrics store (e.g. Thanos) fed by all clusters, and  
- Either one OpenCost querying that store (with cluster labels) or an enterprise federated solution.

## Summary

| Item | In this catalog |
|------|------------------|
| Workload → central OpenCost connection | **Not configured**; they are independent. |
| Data flow | Each cluster: Prometheus (in-cluster) → OpenCost (same cluster). |
| “Central” here | OpenCost on mgmt with UI off (mgmt-only cost). |
| Multi-cluster aggregation | Would require central Prometheus/Thanos + cluster-labeled metrics (and possibly one OpenCost on mgmt pointing at it), or an enterprise federated solution. |
