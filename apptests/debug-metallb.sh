#!/usr/bin/env bash
# Debug MetalLB speaker not ready on the apptests Kind cluster.
# Use when: Helm shows "release installed successfully: metallb/metallb-0.13.7" but
# the metallb-speaker DaemonSet stays at 0/1 ready and "nothing is going on".
# Usage:
#   1. Run tests with cluster left up: SKIP_CLUSTER_TEARDOWN=1 ./catalog-workflow.sh test --appname podinfo
#   2. Then run this from repo root: ./apptests/debug-metallb.sh
#   Or from apptests/: ./debug-metallb.sh
# Cluster name is from the test framework (default: kommanderapptest).

set -e

CLUSTER_NAME="${KIND_CLUSTER_NAME:-kommanderapptest}"
NAMESPACE="metallb-system"
DS_NAME="metallb-speaker"

# Ensure we're in repo root when script is in apptests/
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

if ! kind get kubeconfig --name "$CLUSTER_NAME" &>/dev/null; then
    echo "Kind cluster '$CLUSTER_NAME' not found. List clusters: kind get clusters"
    echo "To keep the cluster after a test run: SKIP_CLUSTER_TEARDOWN=1 ./catalog-workflow.sh test --appname podinfo"
    exit 1
fi

export KUBECONFIG="$(kind get kubeconfig --name "$CLUSTER_NAME")"
echo "=== Using Kind cluster: $CLUSTER_NAME ==="
echo ""

echo "--- Nodes ---"
kubectl get nodes -o wide
echo ""

echo "--- DaemonSet $NAMESPACE/$DS_NAME ---"
kubectl get daemonset -n "$NAMESPACE" "$DS_NAME" -o wide 2>/dev/null || echo "(not found)"
kubectl describe daemonset -n "$NAMESPACE" "$DS_NAME" 2>/dev/null | tail -40
echo ""

echo "--- Secrets in $NAMESPACE (speaker needs metallb-memberlist) ---"
kubectl get secrets -n "$NAMESPACE"
echo ""

echo "--- Pods in $NAMESPACE ---"
kubectl get pods -n "$NAMESPACE" -o wide
echo ""

for pod in $(kubectl get pods -n "$NAMESPACE" -o name 2>/dev/null | head -10); do
    echo "--- Describe $pod ---"
    kubectl describe -n "$NAMESPACE" "$pod" | tail -50
    echo ""
    echo "--- Logs $pod (last 80 lines) ---"
    kubectl logs -n "$NAMESPACE" "$pod" --tail=80 2>/dev/null || true
    echo ""
done

echo "--- Events in $NAMESPACE (last 30) ---"
kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' | tail -30
echo ""

echo "--- All pods not Running (cluster-wide) ---"
kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded 2>/dev/null || true
