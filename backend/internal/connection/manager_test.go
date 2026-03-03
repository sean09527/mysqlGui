package connection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mygui/backend/internal/security"
	"mygui/backend/internal/storage"
	"mygui/backend/types"

	"github.com/vrischmann/userdir"
)

// setupTestManager creates a test ConnectionManager with temporary storage
func setupTestManager(t *testing.T) (*ConnectionManager, *storage.ConfigStorage, func()) {
	// Ensure config directory exists
	configDir := filepath.Join(userdir.GetConfigHome(), "MyGUI")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Use the real ConfigStorage which will create database in user's config directory
	// We'll clean up the test profiles after
	configStorage, err := storage.NewConfigStorage()
	if err != nil {
		t.Fatalf("Failed to create config storage: %v", err)
	}

	encryptor := security.NewEncryptor("test-passphrase-for-testing")
	manager := NewConnectionManager(configStorage, encryptor)

	// Cleanup function - delete all test profiles
	cleanup := func() {
		// List and delete all profiles created during test
		profiles, _ := configStorage.ListProfiles()
		for _, profile := range profiles {
			configStorage.DeleteProfile(profile.ID)
		}
		manager.DisconnectAll()
		configStorage.Close()
	}

	return manager, configStorage, cleanup
}

// TestCreateProfile tests creating a new connection profile
func TestCreateProfile(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Verify profile was created
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	// Verify password was encrypted and then decrypted
	if profiles[0].Password != "testpassword" {
		t.Errorf("Expected password 'testpassword', got '%s'", profiles[0].Password)
	}

	// Verify default values were set
	if profiles[0].Charset != "utf8mb4" {
		t.Errorf("Expected charset 'utf8mb4', got '%s'", profiles[0].Charset)
	}

	if profiles[0].Timeout != 10 {
		t.Errorf("Expected timeout 10, got %d", profiles[0].Timeout)
	}

	// Verify ID was generated
	if profiles[0].ID == "" {
		t.Error("Expected ID to be generated")
	}
}

// TestUpdateProfile tests updating an existing profile
func TestUpdateProfile(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Create initial profile
	profile := types.ConnectionProfile{
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Get the created profile ID
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profileID := profiles[0].ID

	// Update the profile
	updatedProfile := types.ConnectionProfile{
		Name:     "Updated Database",
		Host:     "newhost",
		Port:     3307,
		Username: "newuser",
		Password: "newpassword",
		Database: "newdb",
	}

	err = manager.UpdateProfile(profileID, updatedProfile)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Verify profile was updated
	profiles, err = manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Name != "Updated Database" {
		t.Errorf("Expected name 'Updated Database', got '%s'", profiles[0].Name)
	}

	if profiles[0].Host != "newhost" {
		t.Errorf("Expected host 'newhost', got '%s'", profiles[0].Host)
	}

	if profiles[0].Password != "newpassword" {
		t.Errorf("Expected password 'newpassword', got '%s'", profiles[0].Password)
	}
}

// TestDeleteProfile tests deleting a profile
func TestDeleteProfile(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Create profile
	profile := types.ConnectionProfile{
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Get the created profile ID
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	profileID := profiles[0].ID

	// Delete the profile
	err = manager.DeleteProfile(profileID)
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Verify profile was deleted
	profiles, err = manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 0 {
		t.Fatalf("Expected 0 profiles, got %d", len(profiles))
	}
}

// TestListProfiles tests listing multiple profiles
func TestListProfiles(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Create multiple profiles
	profiles := []types.ConnectionProfile{
		{
			Name:     "Database 1",
			Host:     "host1",
			Port:     3306,
			Username: "user1",
			Password: "pass1",
			Database: "db1",
		},
		{
			Name:     "Database 2",
			Host:     "host2",
			Port:     3307,
			Username: "user2",
			Password: "pass2",
			Database: "db2",
		},
		{
			Name:     "Database 3",
			Host:     "host3",
			Port:     3308,
			Username: "user3",
			Password: "pass3",
			Database: "db3",
		},
	}

	for _, profile := range profiles {
		err := manager.CreateProfile(profile)
		if err != nil {
			t.Fatalf("Failed to create profile: %v", err)
		}
	}

	// List all profiles
	listedProfiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(listedProfiles) != 3 {
		t.Fatalf("Expected 3 profiles, got %d", len(listedProfiles))
	}

	// Verify passwords are decrypted
	for i, profile := range listedProfiles {
		expectedPassword := profiles[i].Password
		if profile.Password != expectedPassword {
			t.Errorf("Profile %d: expected password '%s', got '%s'", i, expectedPassword, profile.Password)
		}
	}
}

// TestCreateProfileWithSSH tests creating a profile with SSH configuration
func TestCreateProfileWithSSH(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		Name:        "SSH Database",
		Host:        "localhost",
		Port:        3306,
		Username:    "testuser",
		Password:    "testpassword",
		Database:    "testdb",
		SSHEnabled:  true,
		SSHHost:     "ssh.example.com",
		SSHPort:     22,
		SSHUsername: "sshuser",
		SSHPassword: "sshpassword",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile with SSH: %v", err)
	}

	// Verify profile was created with SSH settings
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	if !profiles[0].SSHEnabled {
		t.Error("Expected SSH to be enabled")
	}

	if profiles[0].SSHHost != "ssh.example.com" {
		t.Errorf("Expected SSH host 'ssh.example.com', got '%s'", profiles[0].SSHHost)
	}

	// Verify SSH password was encrypted and decrypted
	if profiles[0].SSHPassword != "sshpassword" {
		t.Errorf("Expected SSH password 'sshpassword', got '%s'", profiles[0].SSHPassword)
	}
}

// TestPasswordEncryption tests that passwords are properly encrypted
func TestPasswordEncryption(t *testing.T) {
	manager, storage, cleanup := setupTestManager(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "mysecretpassword",
		Database: "testdb",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Get profile directly from storage to verify encryption
	profiles, err := storage.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles from storage: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	// Password in storage should be encrypted (not equal to original)
	if profiles[0].Password == "mysecretpassword" {
		t.Error("Password should be encrypted in storage")
	}

	// But when retrieved through manager, it should be decrypted
	decryptedProfiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if decryptedProfiles[0].Password != "mysecretpassword" {
		t.Errorf("Expected decrypted password 'mysecretpassword', got '%s'", decryptedProfiles[0].Password)
	}
}

// TestUpdateProfilePreservesCreatedAt tests that updating preserves creation timestamp
func TestUpdateProfilePreservesCreatedAt(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Create profile
	profile := types.ConnectionProfile{
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := manager.CreateProfile(profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Get the created profile
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	originalCreatedAt := profiles[0].CreatedAt
	profileID := profiles[0].ID

	// Wait a bit to ensure timestamps would differ
	time.Sleep(10 * time.Millisecond)

	// Update the profile
	updatedProfile := types.ConnectionProfile{
		Name:     "Updated Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err = manager.UpdateProfile(profileID, updatedProfile)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Verify CreatedAt is preserved
	profiles, err = manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if !profiles[0].CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("CreatedAt should be preserved. Original: %v, Current: %v",
			originalCreatedAt, profiles[0].CreatedAt)
	}

	// Verify UpdatedAt changed
	if profiles[0].UpdatedAt.Equal(originalCreatedAt) {
		t.Error("UpdatedAt should be different from CreatedAt after update")
	}
}

// TestDisconnectAll tests disconnecting all connections
func TestDisconnectAll(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	// Create some mock connections (we can't actually connect to MySQL in tests)
	// So we'll just test that DisconnectAll doesn't error on empty connections
	err := manager.DisconnectAll()
	if err != nil {
		t.Fatalf("DisconnectAll should not error on empty connections: %v", err)
	}
}

// TestBuildDSN tests DSN string construction
func TestBuildDSN(t *testing.T) {
	manager, _, cleanup := setupTestManager(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
		Charset:  "utf8mb4",
		Timeout:  10,
	}

	dsn := manager.buildDSN(profile)

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=true&timeout=10s"
	if dsn != expected {
		t.Errorf("Expected DSN '%s', got '%s'", expected, dsn)
	}
}
