package modules

import (
	"context"
	"errors"
	"fmt"
	"time"

	"opereta/pkg/core"
	"opereta/pkg/inventory"
	"opereta/pkg/utils"
)

// defaultSSHRetryDelay is used if a host does not specify a retry delay.
const defaultSSHRetryDelay = 2 * time.Second

// ShellModule implements the core.Module interface to execute shell commands.
type ShellModule struct{}

// Execute connects via SSH and executes a shell command on the remote host.
func (m ShellModule) Execute(ctx context.Context, host inventory.Host, params map[string]string) (string, error) {
	command, ok := params["command"]
	if !ok {
		return "", errors.New("missing command ('command')")
	}

	port := host.Port
	if port == 0 {
		port = 22
	}

	// Use host-specified SSH retry parameters if provided.
	retryDelay := defaultSSHRetryDelay
	if host.RetrySSH != "" {
		if d, err := time.ParseDuration(host.RetrySSH); err == nil {
			retryDelay = d
		}
	}
	retryCount := 1
	if host.RetrySSHCount > 0 {
		retryCount = host.RetrySSHCount
	}

	client, err := utils.SSHConnect(ctx, host.User, host.Address, host.PrivateKey, host.Password, port, retryDelay, retryCount)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	output, err := utils.RunCommand(ctx, client, command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}
	return output, nil
}

// RegisterModules returns a registry of available modules.
func RegisterModules() map[string]core.Module {
	return map[string]core.Module{
		"shell": ShellModule{},
	}
}
