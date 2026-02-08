package catalogapptests

import (
	"path/filepath"
)

// CatalogApp is an app (e.g. podinfo) from the catalog. Use with cluster.Install(catalogApp).
type CatalogApp struct {
	AppName          string
	VersionToInstall string // empty = latest
}

// Name returns the app name (Helm release name).
func (c *CatalogApp) Name() string {
	return c.AppName
}

// InstallOn implements App. Install the catalog app on the cluster.
func (c *CatalogApp) InstallOn(cluster Cluster) error {
	appPath, err := absolutePathToApp(c.AppName, c.VersionToInstall)
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return cluster.ApplyKustomizations(cluster.Ctx(), helmreleasePath, map[string]string{
		"releaseNamespace": DefaultNamespace,
		"releaseName":      c.AppName,
	})
}

// Install is an alias for InstallOn for backward compatibility.
func (c *CatalogApp) Install(cluster Cluster) error {
	return c.InstallOn(cluster)
}

// InstallPreviousVersion installs the second-to-latest version (for upgrade tests).
func (c *CatalogApp) InstallPreviousVersion(cluster Cluster) error {
	appPath, err := getPrevVersionPath(c.AppName)
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
func (c *CatalogApp) Upgrade(cluster Cluster) error {
	appPath, err := absolutePathToApp(c.AppName, "")
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return cluster.ApplyKustomizations(cluster.Ctx(), helmreleasePath, map[string]string{
		"releaseNamespace": DefaultNamespace,
		"releaseName":      c.AppName,
	})
}
