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
