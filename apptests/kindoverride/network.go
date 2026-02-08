package kind

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/mesosphere/kommander-applications/apptests/docker"
)

var ErrMisconfiguredNetwork = errors.New("misconfigured kind network")

func GetDockerNetworkName() string {
	kindNetwork := "kind"
	if network, ok := os.LookupEnv("KIND_EXPERIMENTAL_DOCKER_NETWORK"); ok {
		kindNetwork = network
	}
	return kindNetwork
}

func EnsureDockerNetworkExist(ctx context.Context, subnet string, internal bool) (*docker.NetworkResource, error) {
	dapi, err := docker.NewAPI()
	if err != nil {
		return nil, fmt.Errorf("could not create docker api:%w", err)
	}
	name := GetDockerNetworkName()
	ok, networkResource, err := dapi.GetNetwork(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker network %s: %w", name, err)
	}
	if ok && networkResource.Internal != internal {
		return nil, fmt.Errorf("internal flag does not match for network %s: %w", name, ErrMisconfiguredNetwork)
	}
	if ok && subnet != "" {
		ipamConfigs := networkResource.IPAM.Config
		if len(ipamConfigs) == 0 {
			return nil, fmt.Errorf("subnet configuration is missing for network %s: %w", name, ErrMisconfiguredNetwork)
		}
		if ipamConfigs[0].Subnet != subnet {
			return nil, fmt.Errorf("subnet expected %s actual %s for network %s: %w", ipamConfigs[0].Subnet, subnet, name, ErrMisconfiguredNetwork)
		}
	}
	if !ok {
		networkResource, err = dapi.CreateNetwork(context.Background(), name, internal, subnet)
		if err != nil {
			return nil, fmt.Errorf("failed to create network %s: %w", name, err)
		}
	}
	return networkResource, nil
}

func EnsureNetworkIsDeleted(ctx context.Context, name string) error {
	dapi, err := docker.NewAPI()
	if err != nil {
		return fmt.Errorf("could not create docker api:%w", err)
	}
	ok, networkResource, err := dapi.GetNetwork(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get docker network %s: %w", name, err)
	}
	if ok {
		if err = dapi.DeleteNetwork(context.Background(), networkResource); err != nil {
			return fmt.Errorf("failed to delete network %s: %w", name, err)
		}
	}
	return nil
}

var kindExperimentalNetworkLock sync.Mutex

func WithKindExperimentalDockerNetwork(networkName string, run func() error) error {
	kindExperimentalNetworkLock.Lock()
	defer kindExperimentalNetworkLock.Unlock()
	defer func(revertTo string) {
		_ = os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", revertTo)
	}(os.Getenv("KIND_EXPERIMENTAL_DOCKER_NETWORK"))
	if err := os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", networkName); err != nil {
		return fmt.Errorf("failed create cluster with network by setting KIND_EXPERIMENTAL_DOCKER_NETWORK=%q: %w", networkName, err)
	}
	return run()
}

func CreateClusterInNetwork(ctx context.Context, clusterName, networkName string) (*Cluster, error) {
	var cluster *Cluster
	err := WithKindExperimentalDockerNetwork(networkName, func() error {
		var err error
		cluster, err = CreateCluster(ctx, clusterName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
