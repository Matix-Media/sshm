package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildConnectionRequiresName(t *testing.T) {
	if _, err := buildConnection("", formValues{Host: "h"}); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestBuildConnectionRequiresHostOrCustom(t *testing.T) {
	if _, err := buildConnection("", formValues{Name: "n"}); err == nil {
		t.Fatal("expected error when host and custom command are both empty")
	}
	if _, err := buildConnection("", formValues{Name: "n", CustomCommand: "mosh box"}); err != nil {
		t.Fatalf("custom command should satisfy host requirement: %v", err)
	}
}

func TestBuildConnectionDefaults(t *testing.T) {
	c, err := buildConnection("", formValues{Name: "n", Host: "h"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.User != "root" || c.Port != "22" || c.Group != "default" {
		t.Fatalf("defaults wrong: user=%q port=%q group=%q", c.User, c.Port, c.Group)
	}
	if c.ID == "" {
		t.Fatal("a new connection should get an id")
	}
}

func TestBuildConnectionKeepsIDOnEdit(t *testing.T) {
	c, err := buildConnection("keep-me", formValues{Name: "n", Host: "h", User: "u", Port: "2222", Group: "g"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "keep-me" {
		t.Fatalf("edit should keep id, got %q", c.ID)
	}
	if c.User != "u" || c.Port != "2222" || c.Group != "g" {
		t.Fatalf("explicit values not preserved: %+v", c)
	}
}

func TestBuildConnectionExpandsIdentityFile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	c, err := buildConnection("", formValues{Name: "n", Host: "h", IdentityFile: "~/.ssh/key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".ssh", "key")
	if c.IdentityFile != want {
		t.Fatalf("identity file = %q, want %q", c.IdentityFile, want)
	}
}
