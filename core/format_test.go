package core

import (
	"strings"
	"testing"
)

func TestConnectionSummary(t *testing.T) {
	cases := []struct {
		c    Connection
		want string
	}{
		{Connection{User: "u", Host: "h", Port: "22"}, "u@h"},
		{Connection{User: "u", Host: "h", Port: "2222"}, "u@h:2222"},
		{Connection{CustomCommand: "mosh me@box"}, "mosh me@box"},
	}
	for _, tc := range cases {
		if got := ConnectionSummary(tc.c); got != tc.want {
			t.Errorf("ConnectionSummary(%+v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestFormatListEmpty(t *testing.T) {
	out := FormatList(nil)
	if !strings.Contains(out, "No connections") {
		t.Fatalf("empty output should hint at adding connections, got %q", out)
	}
}

func TestFormatListTable(t *testing.T) {
	conns := []Connection{
		{ID: "e45a27ab", Name: "VPS 1", Host: "5.75.132.155", User: "matix", Port: "22", Group: "prod"},
		{ID: "0cce1baf", Name: "colima", Host: "127.0.0.1", User: "max", Port: "50738", Group: "imported"},
	}
	out := FormatList(conns)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	if !strings.HasPrefix(lines[0], "ID") || !strings.Contains(lines[0], "CONNECTION") {
		t.Fatalf("missing header row: %q", lines[0])
	}
	// Group-then-name order: "imported" (colima) sorts before "prod" (VPS 1).
	if !strings.Contains(lines[1], "colima") || !strings.Contains(lines[2], "VPS 1") {
		t.Fatalf("rows out of order:\n%s", out)
	}
	// Every row includes its id and summary.
	if !strings.Contains(lines[1], "0cce1baf") || !strings.Contains(lines[1], "max@127.0.0.1:50738") {
		t.Fatalf("colima row incomplete: %q", lines[1])
	}
	// Columns are aligned: the NAME column starts at the same offset on each row.
	if strings.Index(lines[0], "NAME") != strings.Index(lines[1], "colima") {
		t.Fatalf("NAME column misaligned:\n%s", out)
	}
}
