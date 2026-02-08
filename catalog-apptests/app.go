package catalogapptests

import (
	"fmt"
	"path/filepath"
)

// App is something that can be installed on a cluster (e.g. Flux or a catalog app).
type App interface {
	InstallOn(c Cluster) error
}

// Ensure compile-time implementation (only fluxApp implements App; CatalogApp is installed by cluster.Install).
var _ App = fluxApp{}

// fluxApp installs Flux (source, kustomize, helm controllers) on a cluster.
type fluxApp struct{}

// NewFluxApp returns an App that installs Flux on a cluster.
func NewFluxApp() App {
	return fluxApp{}
}

// FluxApp is the default Flux app instance for use in tests (e.g. cluster.Install(FluxApp)).
var FluxApp App = NewFluxApp()

func (fluxApp) InstallOn(cluster Cluster) error {
	impl, ok := cluster.(*clusterImpl)
	if !ok {
		return fmt.Errorf("FluxApp requires cluster from KindCluster")
	}
	return impl.installFlux(impl.ctx)
}

// CatalogApp is an app from the catalog (e.g. podinfo). Use with cluster.Install(catalogApp).
type CatalogApp struct {
	AppName          string
	VersionToInstall string // empty = latest
}

// NewCatalogApp returns a catalog app for the given name and version (version empty = latest).
func NewCatalogApp(appName, versionToInstall string) *CatalogApp {
	return &CatalogApp{AppName: appName, VersionToInstall: versionToInstall}
}

// Name returns the app name (Helm release name).
func (c *CatalogApp) Name() string {
	return c.AppName
}

// Install installs this catalog app on the cluster (cluster.Install pattern).
func (c *CatalogApp) Install(cluster Cluster) error {
	return cluster.Install(c)
}

// InstallPreviousVersion installs the second-to-latest version (for upgrade tests).
// Uses the cluster's Catalog when set; otherwise DefaultCatalog().
func (c *CatalogApp) InstallPreviousVersion(cluster Cluster) error {
	cat := cluster.Catalog()
	if cat == nil {
		var err error
		cat, err = DefaultCatalog()
		if err != nil {
			return err
		}
	}
	appPath, err := cat.PrevVersionPath(c.AppName)
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return cluster.ApplyKustomizations(cluster.Ctx(), helmreleasePath, map[string]string{
		"releaseNamespace": DefaultNamespace,
		"releaseName":      c.AppName,
	})
}

// Upgrade applies the latest version (for upgrade tests).
// Uses the cluster's Catalog when set; otherwise DefaultCatalog().
func (c *CatalogApp) Upgrade(cluster Cluster) error {
	cat := cluster.Catalog()
	if cat == nil {
		var err error
		cat, err = DefaultCatalog()
		if err != nil {
			return err
		}
	}
	appPath, err := cat.PathToApp(c.AppName, "")
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return cluster.ApplyKustomizations(cluster.Ctx(), helmreleasePath, map[string]string{
		"releaseNamespace": DefaultNamespace,
		"releaseName":      c.AppName,
	})
}
