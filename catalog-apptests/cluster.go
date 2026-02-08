package catalogapptests

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/catalog-apptests/framework"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// App is something that can be installed on a cluster (e.g. Flux or a catalog app).
type App interface {
	InstallOn(c Cluster) error
}

// Cluster is a cluster you can install apps on and run assertions against.
type Cluster interface {
	Ctx() context.Context
	Client() ctrlClient.Client
	Install(app App) error
	ApplyKustomizations(ctx context.Context, path string, substitutions map[string]string) error
	Destroy()
}

// NKPManagementCluster is a management cluster that can have workload clusters created from it.
// Workload creation only needs NetworkName() from the management cluster (same Docker network).
// No kubeconfig or other state is read—infra is pluggable via ClusterCreator.
type NKPManagementCluster interface {
	Cluster
	NetworkName() string
	// InstallCentralizedOpencost installs the central/aggregator OpenCost on this management cluster.
	InstallCentralizedOpencost() error
}

// NKPWorkloadCluster is a workload cluster created from an NKP management cluster (CreateFromParent).
// It is a Cluster; the name is for type clarity in tests (e.g. central on mgmt, clients on workloads).
type NKPWorkloadCluster interface {
	Cluster
	// InstallOpencost installs the OpenCost client on this workload cluster.
	InstallOpencost() error
}

// ClusterCreator creates a new cluster on the given network. Used by CreateFromParent.
// Default is Kind; swap to use different infra (e.g. k3d, EKS, etc.).
type ClusterCreator interface {
	CreateCluster(ctx context.Context, networkName, name string) (ClusterHandle, error)
}

// ClusterHandle is the infra-specific handle (kubeconfig path, delete). Implemented by framework.KindCluster.
type ClusterHandle interface {
	KubeconfigFilePath() string
	Delete(ctx context.Context) error
}

// KindCluster creates clusters: Create(ctx, network, name) for mgmt or standalone;
// CreateFromParent(ctx, mgmt, name) for one or more workloads. Mgmt must be an NKPManagementCluster.
var KindCluster = &kindCluster{creator: defaultKindCreator{}}

type kindCluster struct {
	creator ClusterCreator
}

// defaultKindCreator creates Kind clusters on the given Docker network.
type defaultKindCreator struct{}

func (defaultKindCreator) CreateCluster(ctx context.Context, networkName, name string) (ClusterHandle, error) {
	return framework.CreateClusterInNetwork(ctx, name, networkName)
}

// Create creates one cluster on the given network. Use it for mgmt or a standalone cluster.
// The cluster can be used as a parent: pass it to CreateFromParent(ctx, cluster, "workload1"), etc.
func (k *kindCluster) Create(ctx context.Context, network *framework.Network, name string) (Cluster, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if name == "" {
		name = "default"
	}
	if n := os.Getenv("KIND_CLUSTER_NAME"); n != "" {
		name = n
	}

	networkName := "kind"
	if network != nil {
		networkName = network.Name
	}

	var handle ClusterHandle
	var err error
	if network != nil && network.Name != "" && network.Name != "kind" {
		handle, err = k.creator.CreateCluster(ctx, networkName, name)
	} else {
		handle, err = k.createStandalone(ctx, name)
	}
	if err != nil {
		return nil, err
	}

	client, _, err := framework.NewClient(handle.KubeconfigFilePath())
	if err != nil {
		_ = handle.Delete(ctx)
		return nil, err
	}

	c := &clusterImpl{
		name:        name,
		ctx:         ctx,
		handle:      handle,
		client:      client,
		networkName: networkName,
		children:    make(map[string]*clusterImpl),
		destroy:     func() { _ = handle.Delete(ctx) },
	}
	return c, nil
}

func (k *kindCluster) createStandalone(ctx context.Context, name string) (ClusterHandle, error) {
	return framework.CreateCluster(ctx, name)
}

// CreateFromParent creates a new workload cluster on the same network as the management cluster.
// The only thing read from the management cluster is NetworkName(); infra is provided by the creator.
// Call multiple times with different names for workload1, workload2, etc.
func (k *kindCluster) CreateFromParent(ctx context.Context, mgmt NKPManagementCluster, workloadName string) (NKPWorkloadCluster, error) {
	pi, ok := mgmt.(*clusterImpl)
	if !ok {
		return nil, fmt.Errorf("CreateFromParent: management cluster must be from KindCluster.Create")
	}
	if pi.networkName == "" {
		return nil, fmt.Errorf("CreateFromParent: management cluster has no network (cannot create workload)")
	}
	if workloadName == "" {
		workloadName = "workload1"
	}

	pi.mu.Lock()
	if child, ok := pi.children[workloadName]; ok {
		pi.mu.Unlock()
		return child, nil
	}
	pi.mu.Unlock()

	handle, err := k.creator.CreateCluster(ctx, pi.networkName, workloadName)
	if err != nil {
		return nil, err
	}
	workloadClient, _, err := framework.NewClient(handle.KubeconfigFilePath())
	if err != nil {
		_ = handle.Delete(ctx)
		return nil, err
	}

	child := &clusterImpl{
		name:    workloadName,
		ctx:     ctx,
		handle:  handle,
		client:  workloadClient,
		destroy: func() { _ = handle.Delete(ctx) },
	}

	pi.mu.Lock()
	pi.children[workloadName] = child
	// Rebuild destroy to tear down all children then self.
	children := make(map[string]*clusterImpl)
	for n, c := range pi.children {
		children[n] = c
	}
	parentHandle := pi.handle
	pi.destroy = func() {
		for _, c := range children {
			_ = c.handle.Delete(ctx)
		}
		_ = parentHandle.Delete(ctx)
	}
	pi.mu.Unlock()

	return child, nil
}

type clusterImpl struct {
	name        string
	ctx         context.Context
	handle      ClusterHandle
	client      ctrlClient.Client
	isWorkload  bool
	networkName string
	children    map[string]*clusterImpl
	mu          sync.Mutex
	destroy     func()
}

func (c *clusterImpl) Ctx() context.Context                 { return c.ctx }
func (c *clusterImpl) Client() ctrlClient.Client            { return c.client }
func (c *clusterImpl) Destroy()                             { if c.destroy != nil { c.destroy() } }
func (c *clusterImpl) Install(app App) error                 { return app.InstallOn(c) }
func (c *clusterImpl) NetworkName() string                  { return c.networkName }
func (c *clusterImpl) InstallCentralizedOpencost() error    { return c.Install(&CatalogApp{AppName: MulticlusterCentralAppName, VersionToInstall: ""}) }
func (c *clusterImpl) InstallOpencost() error                { return c.Install(&CatalogApp{AppName: MulticlusterTestAppName, VersionToInstall: ""}) }
func (c *clusterImpl) ApplyKustomizations(ctx context.Context, path string, substitutions map[string]string) error {
	return framework.ApplyKustomizations(ctx, c.client, path, substitutions)
}

func (c *clusterImpl) installFlux(ctx context.Context) error {
	return framework.InstallFlux(ctx, c.handle.KubeconfigFilePath(), "")
}

// FluxApp installs Flux (source, kustomize, helm controllers) on a cluster.
var FluxApp App = fluxApp{}

type fluxApp struct{}

func (fluxApp) InstallOn(cluster Cluster) error {
	impl, ok := cluster.(*clusterImpl)
	if !ok {
		return fmt.Errorf("FluxApp requires cluster from KindCluster")
	}
	return impl.installFlux(impl.ctx)
}
