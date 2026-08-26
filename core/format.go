package core

import (
	"fmt"
	"strings"
)

// ConnectionSummary returns the one-line target shown for a connection: the
// custom command if set, otherwise "user@host[:port]".
func ConnectionSummary(c Connection) string {
	if c.CustomCommand != "" {
		return c.CustomCommand
	}
	s := fmt.Sprintf("%s@%s", c.User, c.Host)
	if c.Port != "" && c.Port != "22" {
		s += ":" + c.Port
	}
	return s
}

// FormatList renders connections as an aligned, plain-text table (no color, so
// it pipes cleanly) in the standard group-then-name order. An empty slice yields
// a short hint instead.
func FormatList(conns []Connection) string {
	if len(conns) == 0 {
		return "No connections yet. Run sshm to add one, or import from ~/.ssh/config.\n"
	}
	sorted := append([]Connection(nil), conns...)
	Sort(sorted)

	headers := []string{"ID", "NAME", "CONNECTION", "GROUP"}
	rows := make([][]string, len(sorted))
	for i, c := range sorted {
		rows[i] = []string{c.ID, c.Name, ConnectionSummary(c), c.Group}
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, cell := range r {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder
	writeRow := func(cells []string) {
		for i, cell := range cells {
			if i == len(cells)-1 {
				b.WriteString(cell) // last column needs no trailing padding
			} else {
				fmt.Fprintf(&b, "%-*s  ", widths[i], cell)
			}
		}
		b.WriteByte('\n')
	}

	writeRow(headers)
	for _, r := range rows {
		writeRow(r)
	}
	return b.String()
}
