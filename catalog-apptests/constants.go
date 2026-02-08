package catalogapptests

import "time"

const (
	DefaultNamespace = "default"
	PollInterval     = 2 * time.Second

	// MulticlusterTestAppName is the app installed on workload clusters (client).
	MulticlusterTestAppName = "opencost"
	// MulticlusterCentralAppName is the app installed on mgmt (central/aggregator).
	MulticlusterCentralAppName = "centralized-opencost"
)

// ClusterRole is the NKP role of the cluster; determines which catalog apps are installed (e.g. central on mgmt, client on workload).
type ClusterRole string

const (
	ClusterRoleManagement ClusterRole = "management" // NKPManagementCluster: installs central/aggregator apps
	ClusterRoleWorkload   ClusterRole = "workload" // NKPWorkloadCluster: installs client apps
	ClusterRoleStandalone ClusterRole = "standalone"
)
