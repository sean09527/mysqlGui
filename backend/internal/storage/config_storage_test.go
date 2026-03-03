package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mygui/backend/types"

	"github.com/vrischmann/userdir"
)

// setupTestStorage creates a test ConfigStorage instance
func setupTestStorage(t *testing.T) (*ConfigStorage, func()) {
	// Ensure config directory exists
	configDir := filepath.Join(userdir.GetConfigHome(), "MyGUI")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create ConfigStorage
	storage, err := NewConfigStorage()
	if err != nil {
		t.Fatalf("Failed to create config storage: %v", err)
	}

	// Cleanup function - delete all test profiles
	cleanup := func() {
		profiles, _ := storage.ListProfiles()
		for _, profile := range profiles {
			storage.DeleteProfile(profile.ID)
		}
		storage.Close()
	}

	return storage, cleanup
}

// TestNewConfigStorage tests creating a new ConfigStorage instance
func TestNewConfigStorage(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	if storage == nil {
		t.Fatal("Expected storage to be created")
	}

	if storage.db == nil {
		t.Fatal("Expected database connection to be initialized")
	}
}

// TestSaveProfile tests saving a connection profile
func TestSaveProfile(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		ID:       "test-id-1",
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
		Charset:  "utf8mb4",
		Timeout:  10,
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Verify profile was saved
	savedProfile, err := storage.GetProfile("test-id-1")
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}

	if savedProfile.Name != "Test Database" {
		t.Errorf("Expected name 'Test Database', got '%s'", savedProfile.Name)
	}

	if savedProfile.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", savedProfile.Host)
	}

	if savedProfile.Password != "testpassword" {
		t.Errorf("Expected password 'testpassword', got '%s'", savedProfile.Password)
	}
}

// TestSaveProfileWithSSH tests saving a profile with SSH configuration
func TestSaveProfileWithSSH(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		ID:          "test-ssh-1",
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
		SSHKeyPath:  "/path/to/key",
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save profile with SSH: %v", err)
	}

	// Verify SSH settings were saved
	savedProfile, err := storage.GetProfile("test-ssh-1")
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}

	if !savedProfile.SSHEnabled {
		t.Error("Expected SSH to be enabled")
	}

	if savedProfile.SSHHost != "ssh.example.com" {
		t.Errorf("Expected SSH host 'ssh.example.com', got '%s'", savedProfile.SSHHost)
	}

	if savedProfile.SSHPort != 22 {
		t.Errorf("Expected SSH port 22, got %d", savedProfile.SSHPort)
	}

	if savedProfile.SSHPassword != "sshpassword" {
		t.Errorf("Expected SSH password 'sshpassword', got '%s'", savedProfile.SSHPassword)
	}

	if savedProfile.SSHKeyPath != "/path/to/key" {
		t.Errorf("Expected SSH key path '/path/to/key', got '%s'", savedProfile.SSHKeyPath)
	}
}

// TestUpdateProfile tests updating an existing profile
func TestUpdateProfile(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create initial profile
	profile := types.ConnectionProfile{
		ID:       "test-update-1",
		Name:     "Original Name",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save initial profile: %v", err)
	}

	// Update the profile
	updatedProfile := types.ConnectionProfile{
		ID:       "test-update-1",
		Name:     "Updated Name",
		Host:     "newhost",
		Port:     3307,
		Username: "newuser",
		Password: "newpassword",
		Database: "newdb",
	}

	err = storage.SaveProfile(updatedProfile)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Verify profile was updated
	savedProfile, err := storage.GetProfile("test-update-1")
	if err != nil {
		t.Fatalf("Failed to get updated profile: %v", err)
	}

	if savedProfile.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", savedProfile.Name)
	}

	if savedProfile.Host != "newhost" {
		t.Errorf("Expected host 'newhost', got '%s'", savedProfile.Host)
	}

	if savedProfile.Port != 3307 {
		t.Errorf("Expected port 3307, got %d", savedProfile.Port)
	}
}

// TestGetProfile tests retrieving a profile by ID
func TestGetProfile(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Save a profile
	profile := types.ConnectionProfile{
		ID:       "test-get-1",
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Get the profile
	savedProfile, err := storage.GetProfile("test-get-1")
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}

	if savedProfile.ID != "test-get-1" {
		t.Errorf("Expected ID 'test-get-1', got '%s'", savedProfile.ID)
	}

	if savedProfile.Name != "Test Database" {
		t.Errorf("Expected name 'Test Database', got '%s'", savedProfile.Name)
	}
}

// TestGetProfileNotFound tests getting a non-existent profile
func TestGetProfileNotFound(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := storage.GetProfile("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when getting non-existent profile")
	}
}

// TestListProfiles tests listing all profiles
func TestListProfiles(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create multiple profiles
	profiles := []types.ConnectionProfile{
		{
			ID:       "test-list-1",
			Name:     "Database 1",
			Host:     "host1",
			Port:     3306,
			Username: "user1",
			Password: "pass1",
			Database: "db1",
		},
		{
			ID:       "test-list-2",
			Name:     "Database 2",
			Host:     "host2",
			Port:     3307,
			Username: "user2",
			Password: "pass2",
			Database: "db2",
		},
		{
			ID:       "test-list-3",
			Name:     "Database 3",
			Host:     "host3",
			Port:     3308,
			Username: "user3",
			Password: "pass3",
			Database: "db3",
		},
	}

	for _, profile := range profiles {
		err := storage.SaveProfile(profile)
		if err != nil {
			t.Fatalf("Failed to save profile: %v", err)
		}
	}

	// List all profiles
	listedProfiles, err := storage.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(listedProfiles) != 3 {
		t.Fatalf("Expected 3 profiles, got %d", len(listedProfiles))
	}

	// Verify profiles are sorted by name
	if listedProfiles[0].Name != "Database 1" {
		t.Errorf("Expected first profile name 'Database 1', got '%s'", listedProfiles[0].Name)
	}
}

// TestListProfilesEmpty tests listing when no profiles exist
func TestListProfilesEmpty(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	profiles, err := storage.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profiles) != 0 {
		t.Fatalf("Expected 0 profiles, got %d", len(profiles))
	}
}

// TestDeleteProfile tests deleting a profile
func TestDeleteProfile(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create a profile
	profile := types.ConnectionProfile{
		ID:       "test-delete-1",
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Delete the profile
	err = storage.DeleteProfile("test-delete-1")
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Verify profile was deleted
	_, err = storage.GetProfile("test-delete-1")
	if err == nil {
		t.Fatal("Expected error when getting deleted profile")
	}
}

// TestDeleteProfileNotFound tests deleting a non-existent profile
func TestDeleteProfileNotFound(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	err := storage.DeleteProfile("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent profile")
	}
}

// TestSaveSettings tests saving application settings
func TestSaveSettings(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	err := storage.SaveSettings("theme", "dark")
	if err != nil {
		t.Fatalf("Failed to save setting: %v", err)
	}

	// Verify setting was saved
	value, err := storage.GetSettings("theme")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}

	if value != "dark" {
		t.Errorf("Expected setting value 'dark', got '%s'", value)
	}
}

// TestUpdateSettings tests updating an existing setting
func TestUpdateSettings(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Save initial setting
	err := storage.SaveSettings("language", "en")
	if err != nil {
		t.Fatalf("Failed to save initial setting: %v", err)
	}

	// Update the setting
	err = storage.SaveSettings("language", "zh-CN")
	if err != nil {
		t.Fatalf("Failed to update setting: %v", err)
	}

	// Verify setting was updated
	value, err := storage.GetSettings("language")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}

	if value != "zh-CN" {
		t.Errorf("Expected setting value 'zh-CN', got '%s'", value)
	}
}

// TestGetSettingsNotFound tests getting a non-existent setting
func TestGetSettingsNotFound(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := storage.GetSettings("non-existent-key")
	if err == nil {
		t.Fatal("Expected error when getting non-existent setting")
	}
}

// TestProfileTimestamps tests that timestamps are properly set
func TestProfileTimestamps(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	profile := types.ConnectionProfile{
		ID:       "test-timestamp-1",
		Name:     "Test Database",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpassword",
		Database: "testdb",
	}

	err := storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Get the profile
	savedProfile, err := storage.GetProfile("test-timestamp-1")
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}

	// Verify timestamps are set
	if savedProfile.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if savedProfile.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	originalCreatedAt := savedProfile.CreatedAt

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)

	profile.Name = "Updated Name"
	err = storage.SaveProfile(profile)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Get updated profile
	updatedProfile, err := storage.GetProfile("test-timestamp-1")
	if err != nil {
		t.Fatalf("Failed to get updated profile: %v", err)
	}

	// Verify CreatedAt is preserved
	if !updatedProfile.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should be preserved on update")
	}

	// Verify UpdatedAt changed
	if updatedProfile.UpdatedAt.Equal(originalCreatedAt) {
		t.Error("UpdatedAt should change on update")
	}
}

// TestClose tests closing the storage
func TestClose(t *testing.T) {
	storage, _ := setupTestStorage(t)

	err := storage.Close()
	if err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}

	// Verify database is closed by attempting an operation
	_, err = storage.ListProfiles()
	if err == nil {
		t.Error("Expected error when using closed storage")
	}
}
