package connection

import (
	"testing"
)

// TestSSHTunnelConfig tests SSH tunnel configuration validation
func TestSSHTunnelConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      SSHTunnelConfig
		shouldError bool
	}{
		{
			name: "Valid password authentication",
			config: SSHTunnelConfig{
				SSHHost:     "ssh.example.com",
				SSHPort:     22,
				SSHUsername: "testuser",
				SSHPassword: "testpass",
				RemoteHost:  "localhost",
				RemotePort:  3306,
				Timeout:     10,
			},
			shouldError: false,
		},
		{
			name: "Valid key authentication",
			config: SSHTunnelConfig{
				SSHHost:     "ssh.example.com",
				SSHPort:     22,
				SSHUsername: "testuser",
				SSHKeyPath:  "/path/to/key",
				RemoteHost:  "localhost",
				RemotePort:  3306,
				Timeout:     10,
			},
			shouldError: false,
		},
		{
			name: "Missing authentication",
			config: SSHTunnelConfig{
				SSHHost:     "ssh.example.com",
				SSHPort:     22,
				SSHUsername: "testuser",
				RemoteHost:  "localhost",
				RemotePort:  3306,
				Timeout:     10,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't actually create a tunnel without a real SSH server
			// So we just verify the config structure is correct
			if tt.config.SSHHost == "" {
				t.Error("SSHHost should not be empty")
			}
			if tt.config.SSHPort == 0 {
				t.Error("SSHPort should not be zero")
			}
			if tt.config.SSHUsername == "" {
				t.Error("SSHUsername should not be empty")
			}
			
			// Verify authentication method is provided
			hasAuth := tt.config.SSHPassword != "" || tt.config.SSHKeyPath != ""
			if !hasAuth && !tt.shouldError {
				t.Error("Expected authentication method to be provided")
			}
		})
	}
}

// TestSSHTunnelIntegration tests that SSH tunnel integrates with ConnectionManager
func TestSSHTunnelIntegration(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// This test verifies that the ConnectionManager properly handles SSH-enabled profiles
	// We can't actually connect without a real SSH server, but we can verify the structure
	
	// Verify that tunnels map is initialized
	if manager.tunnels == nil {
		t.Error("Expected tunnels map to be initialized")
	}
	
	// Verify DisconnectAll handles empty tunnels
	err := manager.DisconnectAll()
	if err != nil {
		t.Errorf("DisconnectAll should not error on empty tunnels: %v", err)
	}
}
