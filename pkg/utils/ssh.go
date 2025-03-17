package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConnect establishes an SSH connection to a remote host.
// It retries the connection as specified by retryDelay and retryCount.
func SSHConnect(ctx context.Context, user, address, privateKeyPath, password string, port int, retryDelay time.Duration, retryCount int) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if privateKeyPath != "" {
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parsing private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		return nil, fmt.Errorf("no authentication method provided (private key or password)")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, verify host keys.
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", address, port)
	var lastErr error
	for i := 0; i < retryCount; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("SSH connection canceled: %w", ctx.Err())
		default:
		}
		client, err := ssh.Dial("tcp", addr, config)
		if err == nil {
			return client, nil
		}
		lastErr = err
		// If not the last attempt, wait for retryDelay.
		if i < retryCount-1 {
			time.Sleep(retryDelay)
		}
	}
	return nil, fmt.Errorf("dialing SSH after %d attempts: %w", retryCount, lastErr)
}

// RunCommand executes a command over an established SSH connection.
func RunCommand(ctx context.Context, client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("creating SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("running command '%s': %w", cmd, err)
	}
	return string(output), nil
}
