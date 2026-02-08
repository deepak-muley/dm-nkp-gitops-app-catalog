package catalogapptests

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// AppVersions holds an app name and its version directories (sorted, oldest first).
type AppVersions struct {
	Name     string   // e.g. "podinfo"
	Versions []string // e.g. ["6.9.3", "6.9.4"]
}

// applicationsBase returns the absolute path to the applications/ directory.
// Repo root is inferred from cwd (e.g. catalog-apptests/ or repo root when running go test).
func applicationsBase() (string, error) {
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

// ListCatalogApps returns all catalog apps and their versions by scanning
// applications/<name>/<version>/ directories. Only names that have at least
// one version directory are returned. Versions are sorted (oldest first).
func ListCatalogApps() ([]AppVersions, error) {
	appsDir, err := applicationsBase()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}
	var result []AppVersions
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		appPath := filepath.Join(appsDir, name)
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

// absolutePathToApp returns the absolute path to applications/<app>/<version>.
// If version is empty, returns the latest version (last in sorted order).
func absolutePathToApp(app, version string) (string, error) {
	appsDir, err := applicationsBase()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appsDir, app)
	if version != "" {
		p := filepath.Join(dir, version)
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("no application directory found for app: %s version: %s", app, version)
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
		return "", fmt.Errorf("no application directory found for %s in %s", app, dir)
	}
	return versionDirs[len(versionDirs)-1], nil
}

// getPrevVersionPath returns the path to the second-to-latest version of the app (for upgrade tests).
func getPrevVersionPath(app string) (string, error) {
	appsDir, err := applicationsBase()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(appsDir, app)
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
		return "", fmt.Errorf("no old version found for application: %s", app)
	}
	return versionDirs[len(versionDirs)-2], nil
}
