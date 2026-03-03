package connection

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"mygui/backend/internal/security"
	"mygui/backend/internal/storage"
	"mygui/backend/types"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// ConnectionManager manages database connections and connection profiles
type ConnectionManager struct {
	connections map[string]*sql.DB
	tunnels     map[string]*SSHTunnel
	storage     *storage.ConfigStorage
	encryptor   *security.Encryptor
	mu          sync.RWMutex
}

// NewConnectionManager creates a new ConnectionManager instance
func NewConnectionManager(storage *storage.ConfigStorage, encryptor *security.Encryptor) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*sql.DB),
		tunnels:     make(map[string]*SSHTunnel),
		storage:     storage,
		encryptor:   encryptor,
	}
}

// CreateProfile creates a new connection profile with encrypted password
func (cm *ConnectionManager) CreateProfile(profile types.ConnectionProfile) error {
	// Generate ID if not provided
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}

	// Set default values
	if profile.Charset == "" {
		profile.Charset = "utf8mb4"
	}
	if profile.Timeout == 0 {
		profile.Timeout = 10
	}

	// Encrypt password
	encryptedPassword, err := cm.encryptor.Encrypt(profile.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	profile.Password = encryptedPassword

	// Encrypt SSH password if provided
	if profile.SSHPassword != "" {
		encryptedSSHPassword, err := cm.encryptor.Encrypt(profile.SSHPassword)
		if err != nil {
			return fmt.Errorf("failed to encrypt SSH password: %w", err)
		}
		profile.SSHPassword = encryptedSSHPassword
	}

	// Save to storage
	if err := cm.storage.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateProfile updates an existing connection profile
func (cm *ConnectionManager) UpdateProfile(id string, profile types.ConnectionProfile) error {
	// Verify profile exists
	existingProfile, err := cm.storage.GetProfile(id)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	// Preserve ID
	profile.ID = id

	// Set default values
	if profile.Charset == "" {
		profile.Charset = "utf8mb4"
	}
	if profile.Timeout == 0 {
		profile.Timeout = 10
	}

	// Encrypt password if changed
	if profile.Password != existingProfile.Password {
		encryptedPassword, err := cm.encryptor.Encrypt(profile.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		profile.Password = encryptedPassword
	}

	// Encrypt SSH password if changed
	if profile.SSHPassword != "" && profile.SSHPassword != existingProfile.SSHPassword {
		encryptedSSHPassword, err := cm.encryptor.Encrypt(profile.SSHPassword)
		if err != nil {
			return fmt.Errorf("failed to encrypt SSH password: %w", err)
		}
		profile.SSHPassword = encryptedSSHPassword
	}

	// Preserve creation time
	profile.CreatedAt = existingProfile.CreatedAt

	// Disconnect if currently connected
	cm.mu.Lock()
	if conn, exists := cm.connections[id]; exists {
		conn.Close()
		delete(cm.connections, id)
	}
	cm.mu.Unlock()

	// Save to storage
	if err := cm.storage.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// DeleteProfile deletes a connection profile
func (cm *ConnectionManager) DeleteProfile(id string) error {
	// Disconnect if currently connected
	cm.mu.Lock()
	if conn, exists := cm.connections[id]; exists {
		conn.Close()
		delete(cm.connections, id)
	}
	cm.mu.Unlock()

	// Delete from storage
	if err := cm.storage.DeleteProfile(id); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ListProfiles returns all connection profiles with decrypted passwords
func (cm *ConnectionManager) ListProfiles() ([]types.ConnectionProfile, error) {
	profiles, err := cm.storage.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Decrypt passwords for each profile
	for i := range profiles {
		if profiles[i].Password != "" {
			decryptedPassword, err := cm.encryptor.Decrypt(profiles[i].Password)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt password for profile %s: %w", profiles[i].ID, err)
			}
			profiles[i].Password = decryptedPassword
		}

		if profiles[i].SSHPassword != "" {
			decryptedSSHPassword, err := cm.encryptor.Decrypt(profiles[i].SSHPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt SSH password for profile %s: %w", profiles[i].ID, err)
			}
			profiles[i].SSHPassword = decryptedSSHPassword
		}
	}

	return profiles, nil
}

// TestConnection tests a connection without saving it
func (cm *ConnectionManager) TestConnection(profile types.ConnectionProfile) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(profile.Timeout)*time.Second)
	defer cancel()

	var tunnel *SSHTunnel
	var err error

	// Note: For test connections, passwords are expected to be in plain text
	// They are not encrypted yet since the profile hasn't been saved

	// If SSH is enabled, establish SSH tunnel first
	if profile.SSHEnabled {
		tunnelConfig := SSHTunnelConfig{
			SSHHost:     profile.SSHHost,
			SSHPort:     profile.SSHPort,
			SSHUsername: profile.SSHUsername,
			SSHPassword: profile.SSHPassword, // Use plain text password for testing
			SSHKeyPath:  profile.SSHKeyPath,
			RemoteHost:  profile.Host,
			RemotePort:  profile.Port,
			Timeout:     profile.Timeout,
		}

		tunnel, err = NewSSHTunnel(tunnelConfig)
		if err != nil {
			return fmt.Errorf("SSH tunnel connection failed: %w", err)
		}
		defer tunnel.Close()

		// Update profile to use tunnel's local address
		localPort, err := tunnel.GetLocalPort()
		if err != nil {
			return fmt.Errorf("failed to get tunnel local port: %w", err)
		}

		// Create a copy of profile to avoid modifying the original
		tunnelProfile := profile
		tunnelProfile.Host = "127.0.0.1"
		tunnelProfile.Port = localPort
		profile = tunnelProfile
	}

	// Build DSN (password is plain text for testing)
	dsn := cm.buildDSN(profile)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	// Test connection with timeout
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}

	return nil
}

// Connect establishes a connection to the database using a profile ID
func (cm *ConnectionManager) Connect(profileID string) (*sql.DB, error) {
	// Check if already connected
	cm.mu.RLock()
	if conn, exists := cm.connections[profileID]; exists {
		cm.mu.RUnlock()
		// Test if connection is still alive
		if err := conn.Ping(); err == nil {
			return conn, nil
		}
		// Connection is dead, remove it and clean up tunnel if exists
		cm.mu.Lock()
		delete(cm.connections, profileID)
		if tunnel, exists := cm.tunnels[profileID]; exists {
			tunnel.Close()
			delete(cm.tunnels, profileID)
		}
		cm.mu.Unlock()
	} else {
		cm.mu.RUnlock()
	}

	// Get profile from storage
	profile, err := cm.storage.GetProfile(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Decrypt password
	decryptedPassword, err := cm.encryptor.Decrypt(profile.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}
	profile.Password = decryptedPassword

	var tunnel *SSHTunnel

	// If SSH is enabled, establish SSH tunnel
	if profile.SSHEnabled {
		// Decrypt SSH password if provided
		sshPassword := profile.SSHPassword
		if sshPassword != "" {
			sshPassword, err = cm.encryptor.Decrypt(sshPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt SSH password: %w", err)
			}
		}

		tunnelConfig := SSHTunnelConfig{
			SSHHost:     profile.SSHHost,
			SSHPort:     profile.SSHPort,
			SSHUsername: profile.SSHUsername,
			SSHPassword: sshPassword,
			SSHKeyPath:  profile.SSHKeyPath,
			RemoteHost:  profile.Host,
			RemotePort:  profile.Port,
			Timeout:     profile.Timeout,
		}

		tunnel, err = NewSSHTunnel(tunnelConfig)
		if err != nil {
			return nil, fmt.Errorf("SSH tunnel connection failed: %w", err)
		}

		// Update profile to use tunnel's local address
		localPort, err := tunnel.GetLocalPort()
		if err != nil {
			tunnel.Close()
			return nil, fmt.Errorf("failed to get tunnel local port: %w", err)
		}

		profile.Host = "127.0.0.1"
		profile.Port = localPort
	}

	// Build DSN
	dsn := cm.buildDSN(*profile)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		if tunnel != nil {
			tunnel.Close()
		}
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(profile.Timeout)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		if tunnel != nil {
			tunnel.Close()
		}
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Store connection and tunnel
	cm.mu.Lock()
	cm.connections[profileID] = db
	if tunnel != nil {
		cm.tunnels[profileID] = tunnel
	}
	cm.mu.Unlock()

	return db, nil
}

// Disconnect closes the connection for a profile
func (cm *ConnectionManager) Disconnect(profileID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[profileID]
	if !exists {
		return fmt.Errorf("no active connection for profile: %s", profileID)
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	delete(cm.connections, profileID)

	// Close SSH tunnel if exists
	if tunnel, exists := cm.tunnels[profileID]; exists {
		if err := tunnel.Close(); err != nil {
			return fmt.Errorf("failed to close SSH tunnel: %w", err)
		}
		delete(cm.tunnels, profileID)
	}

	return nil
}

// GetConnection returns an existing connection by profile ID
func (cm *ConnectionManager) GetConnection(profileID string) (*sql.DB, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[profileID]
	if !exists {
		return nil, fmt.Errorf("no active connection for profile: %s", profileID)
	}

	// Test if connection is still alive
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("connection is not alive: %w", err)
	}

	return conn, nil
}

// DisconnectAll closes all active connections
func (cm *ConnectionManager) DisconnectAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errors []error

	// Close all database connections
	for id, conn := range cm.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection %s: %w", id, err))
		}
	}

	// Close all SSH tunnels
	for id, tunnel := range cm.tunnels {
		if err := tunnel.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close SSH tunnel %s: %w", id, err))
		}
	}

	cm.connections = make(map[string]*sql.DB)
	cm.tunnels = make(map[string]*SSHTunnel)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}

// buildDSN builds a MySQL DSN string from a connection profile
func (cm *ConnectionManager) buildDSN(profile types.ConnectionProfile) string {
	// Format: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&timeout=%ds",
		profile.Username,
		profile.Password,
		profile.Host,
		profile.Port,
		profile.Database,
		profile.Charset,
		profile.Timeout,
	)
	return dsn
}
