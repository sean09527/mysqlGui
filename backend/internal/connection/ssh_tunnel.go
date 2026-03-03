package connection

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHTunnel represents an SSH tunnel connection
type SSHTunnel struct {
	sshClient  *ssh.Client
	localAddr  string
	remoteAddr string
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
}

// SSHTunnelConfig contains configuration for establishing an SSH tunnel
type SSHTunnelConfig struct {
	SSHHost     string
	SSHPort     int
	SSHUsername string
	SSHPassword string
	SSHKeyPath  string
	RemoteHost  string
	RemotePort  int
	Timeout     int
}

// NewSSHTunnel creates a new SSH tunnel
func NewSSHTunnel(config SSHTunnelConfig) (*SSHTunnel, error) {
	// Build SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            config.SSHUsername,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Consider implementing proper host key verification
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}

	// Add authentication method
	if config.SSHPassword != "" {
		// Password authentication
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(config.SSHPassword),
		}
	} else if config.SSHKeyPath != "" {
		// Private key authentication
		key, err := loadPrivateKey(config.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(key),
		}
	} else {
		return nil, fmt.Errorf("either SSH password or private key path must be provided")
	}

	// Connect to SSH server
	sshAddr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}

	// Create local listener on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create local listener: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &SSHTunnel{
		sshClient:  sshClient,
		localAddr:  listener.Addr().String(),
		remoteAddr: fmt.Sprintf("%s:%d", config.RemoteHost, config.RemotePort),
		listener:   listener,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start forwarding in background
	go tunnel.forward()

	return tunnel, nil
}

// GetLocalAddr returns the local address (host:port) to connect to
func (t *SSHTunnel) GetLocalAddr() string {
	return t.localAddr
}

// GetLocalPort returns just the local port number
func (t *SSHTunnel) GetLocalPort() (int, error) {
	_, portStr, err := net.SplitHostPort(t.localAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse local address: %w", err)
	}
	
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port: %w", err)
	}
	
	return port, nil
}

// Close closes the SSH tunnel and all connections
func (t *SSHTunnel) Close() error {
	// Cancel context to stop forwarding
	t.cancel()

	// Close listener
	if t.listener != nil {
		t.listener.Close()
	}

	// Close SSH client
	if t.sshClient != nil {
		return t.sshClient.Close()
	}

	return nil
}

// forward handles the port forwarding logic
func (t *SSHTunnel) forward() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			// Accept local connection
			localConn, err := t.listener.Accept()
			if err != nil {
				// Check if context was cancelled
				select {
				case <-t.ctx.Done():
					return
				default:
					// Log error but continue accepting
					continue
				}
			}

			// Handle connection in goroutine
			go t.handleConnection(localConn)
		}
	}
}

// handleConnection handles a single connection through the tunnel
func (t *SSHTunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	// Dial remote connection through SSH tunnel
	remoteConn, err := t.sshClient.Dial("tcp", t.remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	// Copy data bidirectionally
	done := make(chan struct{}, 2)

	// Local -> Remote
	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	// Remote -> Local
	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	// Wait for one direction to finish
	<-done
}

// loadPrivateKey loads a private key from file
func loadPrivateKey(path string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Try to parse the key without passphrase first
	key, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		// If it fails, it might be encrypted
		// For now, we don't support encrypted keys
		return nil, fmt.Errorf("failed to parse private key (encrypted keys are not supported): %w", err)
	}

	return key, nil
}
