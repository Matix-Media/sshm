package tui

import (
	"strings"
	"testing"

	"github.com/Matix-Media/sshm/core"

	tea "github.com/charmbracelet/bubbletea"
)

func sampleApp(t *testing.T) App {
	t.Helper()
	t.Setenv("SSHM_CONFIG_DIR", t.TempDir()) // isolate persistence
	conns := []core.Connection{
		{ID: "0cce1baf", Name: "colima", Host: "127.0.0.1", User: "max", Port: "50738", Group: "imported"},
		{ID: "e45a27ab", Name: "VPS 1 Hetzner", Host: "5.75.132.155", User: "matix", Port: "22", Group: "prod"},
		{ID: "aa112233", Name: "tunnel-box", CustomCommand: "mosh me@box", Group: "default"},
	}
	m, _ := newApp(conns).Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return m.(App)
}

func keyMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestListViewRendersConnections(t *testing.T) {
	v := sampleApp(t).View()
	for _, want := range []string{"sshm", "colima", "VPS 1", "tunnel-box", "prod", "0cce1baf"} {
		if !strings.Contains(v, want) {
			t.Errorf("list view missing %q\n---\n%s", want, v)
		}
	}
}

func TestAddOpensForm(t *testing.T) {
	m, _ := sampleApp(t).Update(keyMsg('a'))
	a := m.(App)
	if a.mode != modeForm {
		t.Fatalf("expected form mode, got %v", a.mode)
	}
	if !strings.Contains(a.View(), "Add Connection") {
		t.Fatalf("add form did not render:\n%s", a.View())
	}
}

func TestEditOpensPrefilledForm(t *testing.T) {
	m, _ := sampleApp(t).Update(keyMsg('e'))
	a := m.(App)
	if a.mode != modeForm {
		t.Fatalf("expected form mode, got %v", a.mode)
	}
	if !strings.Contains(a.View(), "Edit Connection") {
		t.Fatalf("edit form did not render:\n%s", a.View())
	}
}

func TestDeleteOpensConfirm(t *testing.T) {
	m, _ := sampleApp(t).Update(keyMsg('d'))
	a := m.(App)
	if a.mode != modeConfirm {
		t.Fatalf("expected confirm mode, got %v", a.mode)
	}
	if !strings.Contains(a.View(), "Delete connection") {
		t.Fatalf("delete confirm did not render:\n%s", a.View())
	}
}

func TestConnectChoosesSelection(t *testing.T) {
	m, cmd := sampleApp(t).Update(tea.KeyMsg{Type: tea.KeyEnter})
	a := m.(App)
	if a.chosen == nil {
		t.Fatal("enter should choose the highlighted connection")
	}
	if cmd == nil || quitCmd(cmd) == false {
		t.Fatal("enter should also quit the program")
	}
}

func TestQuitReturnsQuitCmd(t *testing.T) {
	_, cmd := sampleApp(t).Update(keyMsg('q'))
	if cmd == nil || !quitCmd(cmd) {
		t.Fatal("q should quit the program")
	}
}

// quitCmd reports whether a command yields tea.Quit.
func quitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestSaveNewConnectionPersists(t *testing.T) {
	a := sampleApp(t)
	m, _ := a.Update(keyMsg('a'))
	a = m.(App)
	// Fill the required fields directly, then submit via ctrl+s.
	a.form.inputs[0].SetValue("New Box")
	a.form.inputs[1].SetValue("9.9.9.9")
	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	a = m.(App)
	if a.mode != modeList {
		t.Fatalf("expected to return to list after save, got %v", a.mode)
	}
	conns, err := core.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, ok := findConn(conns, "New Box"); !ok {
		t.Fatalf("saved connection not persisted: %+v", conns)
	}
}

func findConn(conns []core.Connection, name string) (core.Connection, bool) {
	for _, c := range conns {
		if c.Name == name {
			return c, true
		}
	}
	return core.Connection{}, false
}
