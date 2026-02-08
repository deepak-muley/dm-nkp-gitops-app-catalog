package framework

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/drone/envsubst"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyKustomizations builds the kustomization at path (with substitutions) and applies to the cluster.
func ApplyKustomizations(ctx context.Context, ctrl ctrlClient.Client, path string, substitutions map[string]string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	k := newKustomizer(path, substitutions)
	if err := k.build(); err != nil {
		return err
	}
	out, err := k.output()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(out)
	dec := yaml.NewYAMLOrJSONDecoder(buf, 1<<20)
	for {
		var obj unstructured.Unstructured
		err := dec.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode at %s: %w", path, err)
		}
		if err := ctrl.Patch(ctx, &obj, ctrlClient.Apply, ctrlClient.ForceOwnership, ctrlClient.FieldOwner("catalog-apptests")); err != nil {
			return fmt.Errorf("apply resource: %w", err)
		}
	}
	return nil
}

type kustomizer struct {
	dir    string
	subs   map[string]string
	resmap resmap.ResMap
}

func newKustomizer(dir string, subs map[string]string) *kustomizer {
	if subs == nil {
		subs = make(map[string]string)
	}
	return &kustomizer{dir: dir, subs: subs, resmap: resmap.New()}
}

func (k *kustomizer) build() error {
	opts := krusty.MakeDefaultOptions()
	opts.Reorder = krusty.ReorderOptionLegacy
	kr := krusty.MakeKustomizer(opts)
	rm, err := kr.Run(filesys.MakeFsOnDisk(), k.dir)
	if err != nil {
		return err
	}
	k.resmap.Clear()
	for _, r := range rm.Resources() {
		yamlBytes, err := r.AsYAML()
		if err != nil {
			return err
		}
		substituted, err := envsubst.Eval(string(yamlBytes), func(s string) string {
			return k.subs[s]
		})
		if err != nil {
			return err
		}
		res, err := resourceFromBytes([]byte(substituted))
		if err != nil {
			return err
		}
		k.resmap.Append(res)
	}
	return nil
}

func (k *kustomizer) output() ([]byte, error) {
	return k.resmap.AsYaml()
}

func resourceFromBytes(b []byte) (*resource.Resource, error) {
	fc := resmap.NewFactory(&resource.Factory{})
	rm, err := fc.NewResMapFromBytes(b)
	if err != nil {
		return nil, err
	}
	list := rm.Resources()
	if len(list) == 0 {
		return nil, fmt.Errorf("no resource in yaml")
	}
	return list[0], nil
}
