package framework

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeConfig builds rest.Config from kubeconfig path.
func KubeConfig(kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// NewClient builds a controller-runtime Client and a Kubernetes Clientset from kubeconfig path.
func NewClient(kubeconfigPath string) (ctrlClient.Client, *kubernetes.Clientset, error) {
	cfg, err := KubeConfig(kubeconfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("clientset: %w", err)
	}
	ctrl, err := ctrlClient.New(cfg, ctrlClient.Options{Scheme: NewScheme()})
	if err != nil {
		return nil, nil, fmt.Errorf("controller-runtime client: %w", err)
	}
	return ctrl, clientset, nil
}
