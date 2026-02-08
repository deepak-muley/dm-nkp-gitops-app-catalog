package appscenarios

import (
	"context"
	"path/filepath"

	"github.com/mesosphere/kommander-applications/apptests/environment"

	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/appscenarios/constant"
	"github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/utils"
)

// PodinfoMulticluster installs podinfo to the workload cluster in a multi-cluster env.
// Use with env.ProvisionMultiCluster and env.InstallLatestFluxOnWorkload.
// Verify HelmRelease with env.ClientFor(environment.WorkloadClusterTarget) (or env.WorkloadClient).
type PodinfoMulticluster struct {
	VersionToInstall string // empty = latest
}

// Name returns the app name (Helm release name).
func (p *PodinfoMulticluster) Name() string {
	return "podinfo"
}

// InstallToWorkload applies the podinfo helmrelease to the workload cluster.
// Requires env to be provisioned with ProvisionMultiCluster and Flux installed on workload.
func (p *PodinfoMulticluster) InstallToWorkload(ctx context.Context, env *environment.Env) error {
	appPath, err := utils.AbsolutePathTo(p.Name(), p.VersionToInstall)
	if err != nil {
		return err
	}
	helmreleasePath := filepath.Join(appPath, "helmrelease")
	return env.ApplyKustomizations(ctx, helmreleasePath, map[string]string{
		"releaseNamespace": constant.DEFAULT_NAMESPACE,
		"releaseName":      p.Name(),
	}, environment.WorkloadClusterTarget)
}
