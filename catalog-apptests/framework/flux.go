package framework

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluxcd/cli-utils/pkg/kstatus/polling"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen/install"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen/kustomization"
	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/fluxcd/pkg/ssa"
	"github.com/fluxcd/pkg/ssa/normalize"
	ssautils "github.com/fluxcd/pkg/ssa/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/api/konfig"
)

const fluxNamespace = "kommander-flux"

// InstallFlux installs Flux (source-controller, kustomize-controller, helm-controller) on the cluster.
func InstallFlux(ctx context.Context, kubeconfigPath, namespace string) error {
	log.SetLogger(klog.NewKlogr())
	if namespace == "" {
		namespace = fluxNamespace
	}

	_, clientset, err := NewClient(kubeconfigPath)
	if err != nil {
		return err
	}
	ns := &corev1.Namespace{}
	ns.Name = namespace
	if _, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	options := install.MakeDefaultOptions()
	options.Namespace = namespace
	options.Components = []string{"source-controller", "kustomize-controller", "helm-controller"}

	manifest, err := install.Generate(options, "")
	if err != nil {
		return fmt.Errorf("flux manifest generate: %w", err)
	}
	tmpDir, err := manifestgen.MkdirTempAbs("", namespace)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	if _, err := manifest.WriteFile(tmpDir); err != nil {
		return fmt.Errorf("flux manifest write: %w", err)
	}

	flags := genericclioptions.NewConfigFlags(true)
	flags.KubeConfig = &kubeconfigPath
	opts := &runclient.Options{}

	manifestPath := filepath.Join(tmpDir, manifest.Path)
	objs, err := readFluxObjects(tmpDir, manifestPath)
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		return fmt.Errorf("no objects at %s", manifestPath)
	}
	if err := normalize.UnstructuredList(objs); err != nil {
		return err
	}

	cfg, err := KubeConfig(kubeconfigPath)
	if err != nil {
		return err
	}
	cfg.QPS = opts.QPS
	cfg.Burst = opts.Burst
	mapper, err := flags.ToRESTMapper()
	if err != nil {
		return err
	}
	kubeClient, err := client.New(cfg, client.Options{Mapper: mapper, Scheme: NewScheme()})
	if err != nil {
		return err
	}
	poller := polling.NewStatusPoller(kubeClient, mapper, polling.Options{})
	manager := ssa.NewResourceManager(kubeClient, poller, ssa.Owner{Field: "flux", Group: "fluxcd.io"})

	var stageOne, stageTwo []*unstructured.Unstructured
	for _, u := range objs {
		if ssautils.IsClusterDefinition(u) {
			stageOne = append(stageOne, u)
		} else {
			stageTwo = append(stageTwo, u)
		}
	}
	if len(stageOne) > 0 {
		cs, err := manager.ApplyAll(ctx, stageOne, ssa.DefaultApplyOptions())
		if err != nil {
			return err
		}
		if err := manager.WaitForSet(cs.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: time.Minute}); err != nil {
			return err
		}
	}
	if len(stageTwo) > 0 {
		cs, err := manager.ApplyAll(ctx, stageTwo, ssa.DefaultApplyOptions())
		if err != nil {
			return err
		}
		if err := manager.WaitForSet(cs.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: 5 * time.Minute}); err != nil {
			return err
		}
	}
	return nil
}

// readFluxObjects reads YAML or kustomization from manifestPath (root for kustomize build).
func readFluxObjects(root, manifestPath string) ([]*unstructured.Unstructured, error) {
	fi, err := os.Lstat(manifestPath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() || !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("expected file: %s", manifestPath)
	}
	base := filepath.Base(manifestPath)
	for _, kname := range konfig.RecognizedKustomizationFileNames() {
		if base == kname {
			resources, err := kustomization.BuildWithRoot(root, filepath.Dir(manifestPath))
			if err != nil {
				return nil, err
			}
			return ssautils.ReadObjects(bytes.NewReader(resources))
		}
	}
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ssautils.ReadObjects(f)
}
