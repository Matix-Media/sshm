package core

import (
	"os"
	"path/filepath"
	"strings"
)

// parseSSHConfig parses an ssh_config file into connection profiles, following
// Include directives recursively. seen guards against include cycles. Only
// entries with a resolved HostName are returned.
func parseSSHConfig(path string, seen map[string]bool) []Connection {
	path = ExpandHome(path)
	if seen[path] {
		return nil
	}
	seen[path] = true

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var conns []Connection
	var current []int // indices into conns for the active Host block's aliases

	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := shlexSplit(line)
		if len(parts) == 0 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "include":
			conns = append(conns, parseIncludes(parts[1:], seen)...)
			current = nil
		case "host":
			current = current[:0]
			for _, h := range parts[1:] {
				if strings.ContainsAny(h, "*?") {
					continue // skip wildcard patterns
				}
				conns = append(conns, newImportedConnection(h, path))
				current = append(current, len(conns)-1)
			}
		default:
			for _, idx := range current {
				applyHostField(&conns[idx], key, value)
			}
		}
	}

	return keepWithHost(conns)
}

// parseIncludes resolves each Include pattern (relative patterns are anchored to
// ~/.ssh) and parses every matched file.
func parseIncludes(patterns []string, seen map[string]bool) []Connection {
	var conns []Connection
	for _, pattern := range patterns {
		pattern = ExpandHome(pattern)
		if !filepath.IsAbs(pattern) {
			if home, err := os.UserHomeDir(); err == nil {
				pattern = filepath.Join(home, ".ssh", pattern)
			}
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, m := range matches {
			conns = append(conns, parseSSHConfig(m, seen)...)
		}
	}
	return conns
}

// newImportedConnection builds a fresh profile for a Host entry.
func newImportedConnection(name, source string) Connection {
	return Connection{
		ID:    NewID(),
		Name:  name,
		Port:  "22",
		Group: "imported",
		Notes: "Imported from " + source,
	}
}

// applyHostField fills a connection field from a parsed ssh_config line.
func applyHostField(c *Connection, key, value string) {
	if c == nil {
		return
	}
	switch key {
	case "hostname":
		c.Host = value
	case "user":
		c.User = value
	case "port":
		c.Port = value
	case "identityfile":
		c.IdentityFile = ExpandHome(value)
	}
}

// keepWithHost drops entries that never resolved a HostName.
func keepWithHost(conns []Connection) []Connection {
	out := conns[:0]
	for _, c := range conns {
		if c.Host != "" {
			out = append(out, c)
		}
	}
	return out
}

// mergeNewConnections appends parsed profiles whose name isn't already present
// (case-insensitive) and returns the merged slice plus the number added.
func mergeNewConnections(existing, parsed []Connection) ([]Connection, int) {
	known := make(map[string]bool, len(existing))
	for _, c := range existing {
		known[strings.ToLower(c.Name)] = true
	}
	added := 0
	for _, c := range parsed {
		if known[strings.ToLower(c.Name)] {
			continue
		}
		existing = append(existing, c)
		known[strings.ToLower(c.Name)] = true
		added++
	}
	return existing, added
}

// ImportSSHConfig scans ~/.ssh/config, merges any new profiles into the saved
// connections, and returns how many were added.
func ImportSSHConfig() (int, error) {
	parsed := parseSSHConfig("~/.ssh/config", map[string]bool{})
	existing, err := Load()
	if err != nil {
		return 0, err
	}
	merged, added := mergeNewConnections(existing, parsed)
	if added > 0 {
		if err := Save(merged); err != nil {
			return 0, err
		}
	}
	return added, nil
}
