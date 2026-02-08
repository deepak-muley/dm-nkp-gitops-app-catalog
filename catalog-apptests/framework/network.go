package framework

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// Network holds Docker network identity (name and ID) for Kind clusters.
type Network struct {
	ID   string
	Name string
}

var dockerNetworkMu sync.Mutex

// GetDockerNetworkName returns the network name for Kind (env KIND_EXPERIMENTAL_DOCKER_NETWORK or "kind").
func GetDockerNetworkName() string {
	if n := os.Getenv("KIND_EXPERIMENTAL_DOCKER_NETWORK"); n != "" {
		return n
	}
	return "kind"
}

// EnsureDockerNetworkExist ensures a Docker network exists. subnet and internal can be "" and false for default.
func EnsureDockerNetworkExist(ctx context.Context, subnet string, internal bool) (*Network, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	name := GetDockerNetworkName()
	list, err := cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return nil, fmt.Errorf("network list: %w", err)
	}
	if len(list) > 0 {
		return &Network{ID: list[0].ID, Name: list[0].Name}, nil
	}

	opts := network.CreateOptions{
		Driver:   "bridge",
		Internal: internal,
	}
	if subnet != "" {
		opts.IPAM = &network.IPAM{Config: []network.IPAMConfig{{Subnet: subnet}}}
	}

	resp, err := cli.NetworkCreate(ctx, name, opts)
	if err != nil {
		return nil, fmt.Errorf("network create %s: %w", name, err)
	}
	return &Network{ID: resp.ID, Name: name}, nil
}
