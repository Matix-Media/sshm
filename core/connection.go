// Package core holds the data model and pure logic for sshm: the connection
// store, ssh_config import, identifier resolution, and command building. It has
// no UI and no process-replacing side effects, so it is straightforward to test.
package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Connection is a saved SSH profile. The JSON tags mirror the original Python
// tool exactly so both tools can share ~/.config/sshm/connections.json.
type Connection struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	User          string `json:"user"`
	Port          string `json:"port"`
	IdentityFile  string `json:"identity_file"`
	CustomCommand string `json:"custom_command"`
	Group         string `json:"group"`
	Notes         string `json:"notes"`
}

// ConfigDir returns the config directory, ~/.config/sshm by default. Setting
// SSHM_CONFIG_DIR overrides it (used by tests to avoid touching real data).
func ConfigDir() (string, error) {
	if dir := os.Getenv("SSHM_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sshm"), nil
}

// ConfigFile returns the path to the connections database.
func ConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "connections.json"), nil
}

// ensureDB creates the config dir and an empty database if they don't exist.
func ensureDB() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	file := filepath.Join(dir, "connections.json")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := writeConnections(file, []Connection{}); err != nil {
			return "", err
		}
	}
	return file, nil
}

// Load reads and returns the saved connections, creating the database on first
// use. A missing or unreadable file yields an empty slice.
func Load() ([]Connection, error) {
	file, err := ensureDB()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return []Connection{}, nil
	}
	var conns []Connection
	if err := json.Unmarshal(data, &conns); err != nil {
		return []Connection{}, nil
	}
	return conns, nil
}

// Save persists connections to the shared database.
func Save(conns []Connection) error {
	file, err := ensureDB()
	if err != nil {
		return err
	}
	return writeConnections(file, conns)
}

// writeConnections marshals with 4-space indentation to match the Python tool.
func writeConnections(file string, conns []Connection) error {
	data, err := json.MarshalIndent(conns, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}

// NewID returns an 8-character hex id, replacing Python's uuid4()[:8].
func NewID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b)
}

// Sort orders connections by group then name (case-insensitive), matching the
// original UI ordering.
func Sort(conns []Connection) {
	sort.SliceStable(conns, func(i, j int) bool {
		if conns[i].Group != conns[j].Group {
			return conns[i].Group < conns[j].Group
		}
		return strings.ToLower(conns[i].Name) < strings.ToLower(conns[j].Name)
	})
}
