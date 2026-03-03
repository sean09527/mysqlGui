package types

import "time"

// ConnectionProfile represents a database connection configuration
type ConnectionProfile struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	Username    string     `json:"username"`
	Password    string     `json:"password"`
	Database    string     `json:"database"`
	Charset     string     `json:"charset"`
	Timeout     int        `json:"timeout"`
	SSHEnabled  bool       `json:"sshEnabled"`
	SSHHost     string     `json:"sshHost,omitempty"`
	SSHPort     int        `json:"sshPort,omitempty"`
	SSHUsername string     `json:"sshUsername,omitempty"`
	SSHPassword string     `json:"sshPassword,omitempty"`
	SSHKeyPath  string     `json:"sshKeyPath,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// TestResult represents the result of a connection test
type TestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
