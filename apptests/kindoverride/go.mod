// Replacement for github.com/mesosphere/kommander-applications/apptests/kind;
// package name in this dir is "kind" so the replace works.
module github.com/deepak-muley/dm-nkp-gitops-app-catalog/apptests/kindoverride

go 1.21

require (
	github.com/docker/docker v27.1.1+incompatible
	github.com/mesosphere/kommander-applications/apptests v0.0.0-20251008112838-fac7669f2fdf
	sigs.k8s.io/kind v0.24.0
)
