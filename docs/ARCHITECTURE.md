# Full Stack Architecture

The `dm-nkp-gitops-custom-app` demo can be deployed with the complete observability stack.

## Deployment Order

### Phase 1: Infrastructure

1. **cert-manager** — Certificate management (required for operator webhooks)
2. **letsencrypt-clusterissuer** — Let's Encrypt issuers (optional)
3. **traefik** — Ingress controller (optional)

### Phase 2: Observability Storage

4. **kube-prometheus-stack** — Prometheus + Grafana
5. **loki** — Log aggregation
6. **tempo** — Distributed tracing

### Phase 3: Telemetry Collection

7. **opentelemetry-operator** (recommended) — Manages collectors via CRDs  
   - Apply `OpenTelemetryCollector` CR named `otel` → creates `otel-collector:4317`  
   OR **opentelemetry-collector** — Standalone deployment

### Phase 4: Application

8. **dm-nkp-gitops-custom-app** — Demo app

## Architecture Diagram

```
                              ┌─────────────────────────────────┐
                              │          Internet               │
                              └───────────────┬─────────────────┘
                                              │
                              ┌───────────────▼─────────────────┐
                              │     cert-manager                │
                              │  + letsencrypt-clusterissuer    │
                              └───────────────┬─────────────────┘
                                              │
                              ┌───────────────▼─────────────────┐
                              │      Traefik (Ingress)          │
                              └───────────────┬─────────────────┘
                                              │
┌─────────────────────────────────────────────▼─────────────────────────────┐
│                    dm-nkp-gitops-custom-app (OTLP export)                 │
└─────────────────────────────────────────────┬─────────────────────────────┘
                                              │ OTLP gRPC :4317
                                              ▼
┌───────────────────────────────────────────────────────────────────────────┐
│  OpenTelemetry Operator + Collector (otel-collector)                       │
└─────────────┬───────────────────────────┬────────────────────────┬────────┘
              │ Prometheus Remote Write   │ OTLP                   │ OTLP
              ▼                           ▼                        ▼
┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
│ kube-prometheus-stack│   │ Loki                │   │ Tempo               │
│ (Prometheus+Grafana) │   │ (Logs)              │   │ (Traces)            │
└─────────────────────┘   └─────────────────────┘   └─────────────────────┘
```

## Let's Encrypt

After deploying letsencrypt-clusterissuer, update ACME email in:
`applications/letsencrypt-clusterissuer/1.0.0/helmrelease/clusterissuers.yaml`

```yaml
# Ingress annotation for automatic certs
annotations:
  cert-manager.io/cluster-issuer: "letsencrypt-prod"
```
