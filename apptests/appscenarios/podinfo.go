package appscenarios

import (
	"context"
	"path/filepath"

	"github.com/mesosphere/kommander-applications/apptests/environment"
	scenarios "github.com/mesosphere/kommander-applications/apptests/scenarios"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios/constant"
	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/utils"
)

type Podinfo struct {
	appVersionToInstall string
}

func NewPodinfoScenerio(appVesrionToInstall string) scenarios.AppScenario {
	return &Podinfo{
		appVersionToInstall: appVesrionToInstall,
	}
}

func (pr *Podinfo) Name() string {
	return "podinfo"
}

func (pr *Podinfo) Install(ctx context.Context, env *environment.Env) error {
	appPath, err := utils.AbsolutePathTo(pr.Name(), pr.appVersionToInstall)
	if err != nil {
		return err
	}

	err = pr.install(ctx, env, appPath)
	if err != nil {
		return err
	}

	return err
}

func (pr *Podinfo) install(ctx context.Context, env *environment.Env, appPath string) error {
	// apply the kustomization for the helmrelease
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	err := env.ApplyKustomizations(ctx, helmreleasePath, map[string]string{
		"releaseNamespace": constant.DEFAULT_NAMESPACE,
		"releaseName":      pr.Name(),
	})
	if err != nil {
		return err
	}

	return err
}

func (pr *Podinfo) InstallPreviousVersion(ctx context.Context, env *environment.Env) error {
	appPath, err := utils.GetPrevVAppsUpgradePath(pr.Name())
	if err != nil {
		return err
	}

	err = pr.install(ctx, env, appPath)
	if err != nil {
		return err
	}

	return nil
}

func (pr *Podinfo) Upgrade(ctx context.Context, env *environment.Env) error {
	appPath, err := utils.AbsolutePathTo(pr.Name(), "")
	if err != nil {
		return err
	}

	err = pr.install(ctx, env, appPath)
	if err != nil {
		return err
	}

	return err
}
