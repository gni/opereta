package core

import (
	"context"

	"opereta/pkg/inventory"
)

// Module represents a unit that executes a task on a host.
type Module interface {
	// Execute runs the module logic using the provided context, host, and parameters.
	Execute(ctx context.Context, host inventory.Host, params map[string]string) (string, error)
}
