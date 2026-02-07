package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// absolutePathTo returns the absolute path to the given application directory.
func AbsolutePathTo(application, appVersion string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// determining the execution path.
	var base string
	_, err = os.Stat(filepath.Join(wd, "applications"))
	if os.IsNotExist(err) {
		base = "../.."
	} else {
		base = ""
	}

	dir, err := filepath.Abs(filepath.Join(wd, base, "applications", application))
	if err != nil {
		return "", err
	}

	if appVersion != "" {
		pathToApp := dir + "/" + appVersion
		if !checkIfPathIsValid(pathToApp) {
			return "", fmt.Errorf("no application directory found for app: %s of version: %s", application, appVersion)
		}
		return pathToApp, nil
	}

	// filepath.Glob returns a sorted slice of matching paths
	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf(
			"no application directory found for %s in the given path:%s",
			application, dir)
	}

	return matches[len(matches)-1], nil
}

func checkIfPathIsValid(absPathToApp string) bool {
	if _, err := os.Stat(absPathToApp); err == nil {
		return true
	}
	return false
}

// get one version later from the latest version available for a given application.
func GetPrevVAppsUpgradePath(application string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// determining the execution path.
	var base string
	_, err = os.Stat(filepath.Join(wd, "applications"))
	if os.IsNotExist(err) {
		base = "../.."
	} else {
		base = ""
	}

	dir, err := filepath.Abs(filepath.Join(wd, base, "applications", application))
	if err != nil {
		return "", err
	}

	// filepath.Glob returns a sorted slice of matching paths
	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf(
			"no application directory found for %s in the given path:%s",
			application, dir)
	}

	if len(matches) < 2 {
		return "", fmt.Errorf(
			"no old version found for the application: %s",
			application)
	}

	return matches[len(matches)-2], nil
}
