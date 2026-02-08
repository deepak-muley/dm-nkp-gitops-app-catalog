package framework

import (
	fluxhelmv2 "github.com/fluxcd/helm-controller/api/v2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewScheme returns a runtime.Scheme with Flux CRDs registered (source, kustomize, helm).
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = sourcev1b2.AddToScheme(scheme)
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev1.AddToScheme(scheme)
	_ = fluxhelmv2.AddToScheme(scheme)
	return scheme
}
