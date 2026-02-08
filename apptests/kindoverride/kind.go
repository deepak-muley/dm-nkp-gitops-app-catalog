// Package kind provides function for creating and managing kind clusters.
// This override reads KIND_CLUSTER_NAME when name is empty (instead of defaulting to "kommanderapptest").
package kind

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

//go:embed config/kind.yaml
var kindConfigFile []byte

//go:embed scripts/*
var hackScriptsFS embed.FS

type Cluster struct {
	provider           *cluster.Provider
	kubeconfigFilePath string
	name               string
}

const (
	defaultClusterName             = "kommanderapptest"
	directory_for_kind_hack_scripts = "./tmp-kind-hack-scripts"
)

// CreateCluster creates a new kind cluster with the given name.
// When name is empty, KIND_CLUSTER_NAME env var is used; if unset, defaultClusterName is used.
func CreateCluster(ctx context.Context, name string) (*Cluster, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if name == "" {
		if n := os.Getenv("KIND_CLUSTER_NAME"); n != "" {
			name = n
		} else {
			name = defaultClusterName
		}
	}

	var err error
	err = extractEmbededHackScripts()
	if err != nil {
		return nil, err
	}

	kubeconfigFile, err := os.CreateTemp("", "*-kubeconfig")
	if err != nil {
		return nil, err
	}

	provider := cluster.NewProvider(cluster.ProviderWithLogger(cmd.NewLogger()))
	err = provider.Create(name,
		cluster.CreateWithKubeconfigPath(kubeconfigFile.Name()),
		cluster.CreateWithRawConfig(kindConfigFile),
	)
	if err != nil {
		return nil, err
	}

	err = provider.ExportKubeConfig(name, "", false)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		provider:           provider,
		kubeconfigFilePath: kubeconfigFile.Name(),
		name:               name,
	}, nil
}

// ListNodeNames lists all nodes in the cluster.
func (c *Cluster) ListNodeNames(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	nodes, err := c.provider.ListNodes(c.name)
	if err != nil {
		return nil, err
	}

	nodeNames := make([]string, len(nodes))
	for i, node := range nodes {
		nodeNames[i] = node.String()
	}
	return nodeNames, nil
}

// RunScript runs a script on the given node using docker exec.
func (c *Cluster) RunScript(ctx context.Context, nodeName, script string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer apiClient.Close()

	rst, err := apiClient.ContainerExecCreate(context.Background(), nodeName, container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{script},
	})
	if err != nil {
		return err
	}

	response, err := apiClient.ContainerExecAttach(context.Background(), rst.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer response.Close()

	data, err := io.ReadAll(response.Reader)
	if err != nil {
		return err
	}

	log.Println(string(data))
	return nil
}

// Delete deletes the cluster and the temporary kubeconfig file.
func (c *Cluster) Delete(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := c.provider.Delete(c.name, c.kubeconfigFilePath)
	if err != nil {
		return err
	}
	return os.Remove(c.kubeconfigFilePath)
}

func (c *Cluster) KubeconfigFilePath() string {
	return c.kubeconfigFilePath
}

func (c *Cluster) Name() string {
	return c.name
}

// KubeconfigForPeers generates a kubeconfig file that can be used by other containers
// on the same Docker network to access this cluster.
func (c *Cluster) KubeconfigForPeers() (string, error) {
	kubeconfigData, err := os.ReadFile(c.kubeconfigFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	controlPlaneHost := fmt.Sprintf("%s-control-plane", c.name)
	kubeconfigStr := string(kubeconfigData)
	kubeconfigStr = replaceServerAddress(kubeconfigStr, controlPlaneHost)

	peerKubeconfigFile, err := os.CreateTemp("", "*-peer-kubeconfig")
	if err != nil {
		return "", fmt.Errorf("failed to create peer kubeconfig file: %w", err)
	}
	defer peerKubeconfigFile.Close()

	if _, err := peerKubeconfigFile.WriteString(kubeconfigStr); err != nil {
		return "", err
	}
	return peerKubeconfigFile.Name(), nil
}

func replaceServerAddress(kubeconfig, controlPlaneHost string) string {
	lines := strings.Split(kubeconfig, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "server:") {
			lines[i] = fmt.Sprintf("  server: https://%s:6443", controlPlaneHost)
		}
	}
	return strings.Join(lines, "\n")
}

func extractEmbededHackScripts() error {
	err := os.MkdirAll(directory_for_kind_hack_scripts, 0755)
	if err != nil {
		return err
	}
	entries, err := hackScriptsFS.ReadDir("scripts")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		data, err := hackScriptsFS.ReadFile("scripts/" + entry.Name())
		if err != nil {
			return err
		}
		targetPath := filepath.Join(directory_for_kind_hack_scripts, entry.Name())
		err = os.WriteFile(targetPath, data, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
