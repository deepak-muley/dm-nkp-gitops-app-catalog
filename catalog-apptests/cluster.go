package catalogapptests

// High-level object model:
//   - Network (framework.Network): Docker network clusters are created on.
//   - Catalog: dm-nkp-gitops-app-catalog (applications/ discovery); used by Cluster to resolve and install catalog apps.
//   - Cluster: uses Network and Catalog; has a Role (management | workload | standalone). Install behavior is per role:
//     NKPManagementCluster (role=management): InstallCentralizedOpencost, CreateFromParent for workloads.
//     NKPWorkloadCluster (role=workload): InstallOpencost.
//   - App: installable unit (FluxApp, CatalogApp); Cluster.Install(app) dispatches by type.
// ClusterConfig binds Network + optional Catalog + Name when creating a cluster.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/catalog-apptests/framework"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster is a cluster that uses a Network and optionally a Catalog, and installs apps by Role.
// Role determines which catalog apps are appropriate (e.g. InstallCentralizedOpencost on management, InstallOpencost on workload).
type Cluster interface {
	Ctx() context.Context
	Client() ctrlClient.Client
	Catalog() Catalog
	Network() *framework.Network
	// Role is the NKP cluster role (management, workload, standalone); install behavior is per role.
	Role() ClusterRole
	Install(app interface{}) error
	ApplyKustomizations(ctx context.Context, path string, substitutions map[string]string) error
	Destroy()
}

// ClusterConfig configures a new cluster: it uses the given Network and optional Catalog.
type ClusterConfig struct {
	Network *framework.Network
	Catalog Catalog // optional; nil => DefaultCatalog() used when installing catalog apps
	Name    string
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
	return framework.NewKindClusterInNetwork(ctx, name, networkName)
}

// Create creates one cluster from config (uses config.Network and config.Catalog). Use for mgmt or standalone.
// The cluster can be used as a parent: pass it to CreateFromParent(ctx, cluster, "workload1"), etc.
func (k *kindCluster) Create(ctx context.Context, config ClusterConfig) (Cluster, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	name := config.Name
	if name == "" {
		name = "default"
	}
	if n := os.Getenv("KIND_CLUSTER_NAME"); n != "" {
		name = n
	}

	networkName := "kind"
	if config.Network != nil {
		networkName = config.Network.Name
	}

	var handle ClusterHandle
	var err error
	if config.Network != nil && config.Network.Name != "" && config.Network.Name != "kind" {
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

	role := ClusterRoleStandalone
	if config.Network != nil && config.Network.Name != "" && config.Network.Name != "kind" {
		role = ClusterRoleManagement
	}
	c := &clusterImpl{
		name:        name,
		ctx:         ctx,
		handle:      handle,
		client:      client,
		catalog:     config.Catalog,
		network:     config.Network,
		networkName: networkName,
		role:        role,
		children:    make(map[string]*clusterImpl),
		destroy:     func() { _ = handle.Delete(ctx) },
	}
	return c, nil
}

func (k *kindCluster) createStandalone(ctx context.Context, name string) (ClusterHandle, error) {
	return framework.NewKindCluster(ctx, name)
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

	parentCatalog := pi.Catalog()
	child := &clusterImpl{
		name:        workloadName,
		ctx:         ctx,
		handle:      handle,
		client:      workloadClient,
		catalog:     parentCatalog,
		network:     pi.network,
		networkName: pi.networkName,
		role:        ClusterRoleWorkload,
		destroy:     func() { _ = handle.Delete(ctx) },
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
	catalog     Catalog
	network     *framework.Network
	networkName string
	role        ClusterRole
	children    map[string]*clusterImpl
	mu          sync.Mutex
	destroy     func()
}

func (c *clusterImpl) Ctx() context.Context      { return c.ctx }
func (c *clusterImpl) Client() ctrlClient.Client { return c.client }
func (c *clusterImpl) Catalog() Catalog             { return c.catalog }
func (c *clusterImpl) Network() *framework.Network { return c.network }
func (c *clusterImpl) Role() ClusterRole           { return c.role }
func (c *clusterImpl) Destroy()                    { if c.destroy != nil { c.destroy() } }

func (c *clusterImpl) Install(app interface{}) error {
	switch a := app.(type) {
	case App:
		return a.InstallOn(c)
	case *CatalogApp:
		return c.installCatalogApp(a)
	default:
		return fmt.Errorf("unsupported app type: %T", app)
	}
}

func (c *clusterImpl) installCatalogApp(app *CatalogApp) error {
	cat := c.catalog
	if cat == nil {
		var err error
		cat, err = DefaultCatalog()
		if err != nil {
			return err
		}
	}
	appPath, err := cat.PathToApp(app.AppName, app.VersionToInstall)
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return c.ApplyKustomizations(c.ctx, helmreleasePath, map[string]string{
		"releaseNamespace": DefaultNamespace,
		"releaseName":      app.AppName,
	})
}

func (c *clusterImpl) NetworkName() string                  { return c.networkName }
func (c *clusterImpl) InstallCentralizedOpencost() error { return c.Install(NewCatalogApp(MulticlusterCentralAppName, "")) }
func (c *clusterImpl) InstallOpencost() error            { return c.Install(NewCatalogApp(MulticlusterTestAppName, "")) }
func (c *clusterImpl) ApplyKustomizations(ctx context.Context, path string, substitutions map[string]string) error {
	return framework.ApplyKustomizations(ctx, c.client, path, substitutions)
}

func (c *clusterImpl) installFlux(ctx context.Context) error {
	return framework.InstallFlux(ctx, c.handle.KubeconfigFilePath(), "")
}
