package catalogapptests

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// AppVersions holds an app name and its version directories (sorted, oldest first).
type AppVersions struct {
	Name     string   // e.g. "podinfo"
	Versions []string // e.g. ["6.9.3", "6.9.4"]
}

// Catalog provides discovery and iteration over applications/<name>/<version>/.
// Use NewCatalog() to create one; path resolution uses the applications/ base directory.
type Catalog interface {
	// Apps returns all catalog apps and their versions (sorted by name; versions oldest first).
	Apps() ([]AppVersions, error)
	// Each calls f for each app; stops on first error and returns it.
	Each(f func(AppVersions) error) error
	// PathToApp returns the absolute path to applications/<app>/<version>. Empty version = latest.
	PathToApp(appName, version string) (string, error)
	// PrevVersionPath returns the path to the second-to-latest version (for upgrade tests).
	PrevVersionPath(appName string) (string, error)
}

var _ Catalog = (*catalog)(nil)

type catalog struct {
	basePath string
}

// NewCatalog discovers the applications/ directory from cwd (repo root or catalog-apptests/)
// and returns a Catalog. Use it for listing and path resolution; prefer one instance per test run.
func NewCatalog() (Catalog, error) {
	base, err := resolveApplicationsBase()
	if err != nil {
		return nil, err
	}
	return &catalog{basePath: base}, nil
}

func resolveApplicationsBase() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, base := range []string{wd, filepath.Join(wd, ".."), filepath.Join(wd, "..", "..")} {
		appsPath := filepath.Join(base, "applications")
		if st, err := os.Stat(appsPath); err == nil && st.IsDir() {
			return filepath.Abs(appsPath)
		}
	}
	return "", os.ErrNotExist
}

func isVersionDir(path string) bool {
	if st, err := os.Stat(filepath.Join(path, "helmrelease")); err == nil && st.IsDir() {
		return true
	}
	if st, err := os.Stat(filepath.Join(path, "metadata.yaml")); err == nil && !st.IsDir() {
		return true
	}
	return false
}

func (c *catalog) Apps() ([]AppVersions, error) {
	entries, err := os.ReadDir(c.basePath)
	if err != nil {
		return nil, err
	}
	var result []AppVersions
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		appPath := filepath.Join(c.basePath, name)
		subs, err := os.ReadDir(appPath)
		if err != nil {
			continue
		}
		var versions []string
		for _, s := range subs {
			if !s.IsDir() {
				continue
			}
			vPath := filepath.Join(appPath, s.Name())
			if isVersionDir(vPath) {
				versions = append(versions, s.Name())
			}
		}
		if len(versions) == 0 {
			continue
		}
		sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
		result = append(result, AppVersions{Name: name, Versions: versions})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Each implements Catalog: iterates over apps and calls f; stops on first error.
func (c *catalog) Each(f func(AppVersions) error) error {
	apps, err := c.Apps()
	if err != nil {
		return err
	}
	for _, av := range apps {
		if err := f(av); err != nil {
			return err
		}
	}
	return nil
}

func (c *catalog) PathToApp(appName, version string) (string, error) {
	dir := filepath.Join(c.basePath, appName)
	if version != "" {
		p := filepath.Join(dir, version)
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("no application directory found for app: %s version: %s", appName, version)
		}
		return p, nil
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}
	var versionDirs []string
	for _, m := range matches {
		if st, err := os.Stat(m); err == nil && st.IsDir() && isVersionDir(m) {
			versionDirs = append(versionDirs, m)
		}
	}
	sort.Strings(versionDirs)
	if len(versionDirs) == 0 {
		return "", fmt.Errorf("no application directory found for %s in %s", appName, dir)
	}
	return versionDirs[len(versionDirs)-1], nil
}

func (c *catalog) PrevVersionPath(appName string) (string, error) {
	dir := filepath.Join(c.basePath, appName)
	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}
	var versionDirs []string
	for _, m := range matches {
		if st, err := os.Stat(m); err == nil && st.IsDir() && isVersionDir(m) {
			versionDirs = append(versionDirs, m)
		}
	}
	sort.Strings(versionDirs)
	if len(versionDirs) < 2 {
		return "", fmt.Errorf("no old version found for application: %s", appName)
	}
	return versionDirs[len(versionDirs)-2], nil
}

// default catalog for package-level helpers (lazy init)
var (
	defaultCatalog    Catalog
	defaultCatalogMu  sync.Once
	defaultCatalogErr error
)

// DefaultCatalog returns a lazily-initialized Catalog (applications/ from cwd).
// Used by ListCatalogApps, absolutePathToApp, getPrevVersionPath for backward compatibility.
func DefaultCatalog() (Catalog, error) {
	defaultCatalogMu.Do(func() {
		defaultCatalog, defaultCatalogErr = NewCatalog()
	})
	return defaultCatalog, defaultCatalogErr
}

// ListCatalogApps returns all catalog apps and their versions (uses DefaultCatalog).
func ListCatalogApps() ([]AppVersions, error) {
	cat, err := DefaultCatalog()
	if err != nil {
		return nil, err
	}
	return cat.Apps()
}
