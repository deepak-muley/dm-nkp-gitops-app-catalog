package framework

import (
	"context"
	"fmt"
	"os"
	"sync"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

// KindCluster is a Kind cluster (name, kubeconfig path, provider for delete).
type KindCluster struct {
	name       string
	kubeconfig string
	provider   *cluster.Provider
}

// Name returns the cluster name.
func (k *KindCluster) Name() string { return k.name }

// KubeconfigFilePath returns the kubeconfig path.
func (k *KindCluster) KubeconfigFilePath() string { return k.kubeconfig }

// Delete removes the cluster and the temp kubeconfig file.
func (k *KindCluster) Delete(ctx context.Context) error {
	if k.provider == nil {
		return nil
	}
	err := k.provider.Delete(k.name, k.kubeconfig)
	if rmErr := os.Remove(k.kubeconfig); rmErr != nil && err == nil {
		err = rmErr
	}
	return err
}

// Minimal Kind config: one control-plane node, pod subnet for compatibility.
var defaultKindConfig = []byte(`
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  podSubnet: "172.16.0.0/16"
nodes:
  - role: control-plane
`)

var kindCreateMu sync.Mutex

// CreateClusterInNetwork creates a Kind cluster in the given Docker network.
// It sets KIND_EXPERIMENTAL_DOCKER_NETWORK for the duration of create.
func CreateClusterInNetwork(ctx context.Context, clusterName, networkName string) (*KindCluster, error) {
	kindCreateMu.Lock()
	defer kindCreateMu.Unlock()

	restore := os.Getenv("KIND_EXPERIMENTAL_DOCKER_NETWORK")
	defer func() { _ = os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", restore) }()
	if err := os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", networkName); err != nil {
		return nil, fmt.Errorf("set KIND_EXPERIMENTAL_DOCKER_NETWORK: %w", err)
	}

	return CreateCluster(ctx, clusterName)
}

// CreateCluster creates a Kind cluster (uses default network if not set via env).
func CreateCluster(ctx context.Context, name string) (*KindCluster, error) {
	if name == "" {
		name = "catalog-test"
	}
	kubeconfigFile, err := os.CreateTemp("", "*-kubeconfig")
	if err != nil {
		return nil, err
	}
	kubeconfigPath := kubeconfigFile.Name()
	_ = kubeconfigFile.Close()

	provider := cluster.NewProvider(cluster.ProviderWithLogger(cmd.NewLogger()))
	err = provider.Create(name,
		cluster.CreateWithKubeconfigPath(kubeconfigPath),
		cluster.CreateWithRawConfig(defaultKindConfig),
	)
	if err != nil {
		_ = os.Remove(kubeconfigPath)
		return nil, err
	}
	if err := provider.ExportKubeConfig(name, "", false); err != nil {
		_ = provider.Delete(name, kubeconfigPath)
		_ = os.Remove(kubeconfigPath)
		return nil, err
	}
	return &KindCluster{name: name, kubeconfig: kubeconfigPath, provider: provider}, nil
}
