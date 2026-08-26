package core

import (
	"fmt"
	"strings"
)

// Resolve finds a connection by id first, then by name (case-insensitive). It
// returns an error naming the reference when no match is found.
func Resolve(conns []Connection, ref string) (Connection, error) {
	for _, c := range conns {
		if c.ID == ref {
			return c, nil
		}
	}
	for _, c := range conns {
		if strings.EqualFold(c.Name, ref) {
			return c, nil
		}
	}
	return Connection{}, fmt.Errorf("no connection matching %q", ref)
}

// BuildCommand returns the argv used to launch a connection, appending any extra
// arguments (e.g. a remote command). A custom command, when set, takes
// precedence over the host/user/port fields — matching the original behavior.
func BuildCommand(c Connection, extra []string) []string {
	if c.CustomCommand != "" {
		return append(shlexSplit(c.CustomCommand), extra...)
	}

	argv := []string{"ssh"}
	if c.Port != "" && c.Port != "22" {
		argv = append(argv, "-p", c.Port)
	}
	if c.IdentityFile != "" {
		argv = append(argv, "-i", ExpandHome(c.IdentityFile))
	}
	argv = append(argv, fmt.Sprintf("%s@%s", c.User, c.Host))
	return append(argv, extra...)
}
