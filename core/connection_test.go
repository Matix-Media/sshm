package core

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func withTempConfig(t *testing.T) {
	t.Helper()
	t.Setenv("SSHM_CONFIG_DIR", t.TempDir())
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	withTempConfig(t)
	conns, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(conns) != 0 {
		t.Fatalf("want empty, got %d", len(conns))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempConfig(t)
	in := []Connection{{ID: "abc", Name: "Box", Host: "h", User: "u", Port: "22"}}
	if err := Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(out) != 1 || out[0] != in[0] {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}

func TestSaveUsesFourSpaceIndent(t *testing.T) {
	withTempConfig(t)
	if err := Save([]Connection{{ID: "x", Name: "n"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	file, _ := ConfigFile()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), "\n    {") || !strings.Contains(string(data), "\n        \"id\"") {
		t.Fatalf("expected 4-space indent, got:\n%s", data)
	}
}

func TestEnsureDBCreatesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SSHM_CONFIG_DIR", dir)
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "connections.json")); err != nil {
		t.Fatalf("db not created: %v", err)
	}
}

func TestNewIDIsEightHexChars(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}$`)
	for i := 0; i < 100; i++ {
		if id := NewID(); !re.MatchString(id) {
			t.Fatalf("bad id: %q", id)
		}
	}
}

func TestSortByGroupThenName(t *testing.T) {
	conns := []Connection{
		{Name: "zeta", Group: "b"},
		{Name: "Beta", Group: "a"},
		{Name: "alpha", Group: "a"},
	}
	Sort(conns)
	got := []string{conns[0].Name, conns[1].Name, conns[2].Name}
	want := []string{"alpha", "Beta", "zeta"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sort = %v, want %v", got, want)
		}
	}
}
