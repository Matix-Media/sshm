package core

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func findByName(conns []Connection, name string) (Connection, bool) {
	for _, c := range conns {
		if c.Name == name {
			return c, true
		}
	}
	return Connection{}, false
}

func TestParseSSHConfigFields(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
# a comment
Host web
    HostName 10.0.0.1
    User deploy
    Port 2222
    IdentityFile ~/.ssh/web_key

Host *
    User ignored

Host db shadow
    HostName 10.0.0.2
`)
	conns := parseSSHConfig(cfg, map[string]bool{})
	if len(conns) != 3 {
		t.Fatalf("want 3 conns (web, db, shadow), got %d: %+v", len(conns), conns)
	}
	web, ok := findByName(conns, "web")
	if !ok || web.Host != "10.0.0.1" || web.User != "deploy" || web.Port != "2222" {
		t.Fatalf("web parsed wrong: %+v", web)
	}
	if web.IdentityFile == "~/.ssh/web_key" || web.IdentityFile == "" {
		t.Fatalf("identity file should be expanded, got %q", web.IdentityFile)
	}
	// Both names of a multi-host line share the same HostName.
	if _, ok := findByName(conns, "shadow"); !ok {
		t.Fatal("multi-host second name missing")
	}
}

func TestParseSSHConfigSkipsWildcard(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, "Host *\n    User x\n    HostName y\n")
	if conns := parseSSHConfig(cfg, map[string]bool{}); len(conns) != 0 {
		t.Fatalf("wildcard host should be skipped, got %+v", conns)
	}
}

func TestParseSSHConfigDropsHostless(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, "Host nohost\n    User x\n")
	if conns := parseSSHConfig(cfg, map[string]bool{}); len(conns) != 0 {
		t.Fatalf("entry without HostName should be dropped, got %+v", conns)
	}
}

func TestParseSSHConfigInclude(t *testing.T) {
	dir := t.TempDir()
	inc := filepath.Join(dir, "extra.conf")
	writeFile(t, inc, "Host included\n    HostName 1.2.3.4\n")
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, "Include "+inc+"\nHost main\n    HostName 5.6.7.8\n")

	conns := parseSSHConfig(cfg, map[string]bool{})
	if _, ok := findByName(conns, "included"); !ok {
		t.Fatalf("include not parsed: %+v", conns)
	}
	if _, ok := findByName(conns, "main"); !ok {
		t.Fatalf("main host missing after include: %+v", conns)
	}
}

func TestParseSSHConfigIncludeCycle(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	writeFile(t, a, "Include "+b+"\nHost hosta\n    HostName 1.1.1.1\n")
	writeFile(t, b, "Include "+a+"\nHost hostb\n    HostName 2.2.2.2\n")
	// Should terminate despite the cycle.
	conns := parseSSHConfig(a, map[string]bool{})
	if _, ok := findByName(conns, "hosta"); !ok {
		t.Fatal("hosta missing")
	}
	if _, ok := findByName(conns, "hostb"); !ok {
		t.Fatal("hostb missing")
	}
}

func TestMergeNewConnectionsDedupes(t *testing.T) {
	existing := []Connection{{Name: "web"}}
	parsed := []Connection{{Name: "WEB"}, {Name: "new"}}
	merged, added := mergeNewConnections(existing, parsed)
	if added != 1 {
		t.Fatalf("want 1 added, got %d", added)
	}
	if len(merged) != 2 {
		t.Fatalf("want 2 total, got %d", len(merged))
	}
}
