package constant

import "time"

const (
	DEFAULT_NAMESPACE = "default" // TODO: we need to create a namespace dynamically during runtime
	POLL_INTERVAL     = 2 * time.Second
)
