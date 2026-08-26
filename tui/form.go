package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Matix-Media/sshm/core"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// formAction is the outcome of handling a key in the form.
type formAction int

const (
	formNone formAction = iota
	formSubmit
	formCancel
)

// formValues holds the raw string inputs; kept separate from the textinput
// widgets so validation/mapping is pure and unit-testable.
type formValues struct {
	Name, Host, User, Port, IdentityFile, CustomCommand, Group, Notes string
}

// buildConnection validates raw form values and produces a Connection. Name is
// always required; a host is required unless a custom command is provided.
// Empty user/port/group fall back to sensible defaults, matching the original
// tool. When id is empty a new one is generated (add); otherwise it is kept
// (edit).
func buildConnection(id string, v formValues) (core.Connection, error) {
	name := strings.TrimSpace(v.Name)
	if name == "" {
		return core.Connection{}, errors.New("name is required")
	}
	host := strings.TrimSpace(v.Host)
	custom := strings.TrimSpace(v.CustomCommand)
	if host == "" && custom == "" {
		return core.Connection{}, errors.New("host is required (or set a custom command)")
	}

	user := orDefault(v.User, "root")
	port := orDefault(v.Port, "22")
	group := orDefault(v.Group, "default")
	if id == "" {
		id = core.NewID()
	}

	return core.Connection{
		ID:            id,
		Name:          name,
		Host:          host,
		User:          user,
		Port:          port,
		IdentityFile:  core.ExpandHome(strings.TrimSpace(v.IdentityFile)),
		CustomCommand: custom,
		Group:         group,
		Notes:         strings.TrimSpace(v.Notes),
	}, nil
}

func orDefault(s, def string) string {
	if s = strings.TrimSpace(s); s != "" {
		return s
	}
	return def
}

var formLabels = []string{"Name", "Host", "User", "Port", "Identity File", "Custom Command", "Group", "Notes"}

// Form is the add/edit view: a set of focus-cycling text inputs.
type Form struct {
	title  string
	id     string // empty when adding
	inputs []textinput.Model
	focus  int
	err    string
}

func newAddForm(width int) *Form {
	return newForm("Add Connection", "", core.Connection{}, width)
}

func newEditForm(c core.Connection, width int) *Form {
	return newForm("Edit Connection", c.ID, c, width)
}

func newForm(title, id string, c core.Connection, width int) *Form {
	placeholders := []string{
		"My Server", "10.0.0.1 or example.com", "root", "22",
		"~/.ssh/id_ed25519 (optional)", "optional; overrides host/user/port",
		"default", "optional",
	}
	values := []string{c.Name, c.Host, c.User, c.Port, c.IdentityFile, c.CustomCommand, c.Group, c.Notes}

	inputs := make([]textinput.Model, len(formLabels))
	for i := range inputs {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = placeholders[i]
		ti.SetValue(values[i])
		inputs[i] = ti
	}
	f := &Form{title: title, id: id, inputs: inputs}
	f.setWidth(width)
	f.inputs[0].Focus()
	return f
}

// setWidth sizes the input fields to the terminal.
func (f *Form) setWidth(width int) {
	w := width - 20
	if w < 20 {
		w = 20
	}
	if w > 60 {
		w = 60
	}
	for i := range f.inputs {
		f.inputs[i].Width = w
	}
}

// focusCmd returns the cursor-blink command for the focused field.
func (f *Form) focusCmd() tea.Cmd { return textinput.Blink }

// Update handles a message and reports whether the form was submitted or
// cancelled.
func (f *Form) Update(msg tea.Msg) (formAction, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return formNone, f.updateFocused(msg)
	}
	switch {
	case key.Matches(km, keys.Cancel):
		return formCancel, nil
	case key.Matches(km, keys.Save):
		return formSubmit, nil
	case km.Type == tea.KeyEnter:
		if f.focus == len(f.inputs)-1 {
			return formSubmit, nil
		}
		f.advance(1)
		return formNone, textinput.Blink
	case key.Matches(km, keys.Next):
		f.advance(1)
		return formNone, textinput.Blink
	case key.Matches(km, keys.Prev):
		f.advance(-1)
		return formNone, textinput.Blink
	}
	return formNone, f.updateFocused(msg)
}

func (f *Form) updateFocused(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return cmd
}

func (f *Form) advance(delta int) {
	f.inputs[f.focus].Blur()
	n := len(f.inputs)
	f.focus = (f.focus + delta + n) % n
	f.inputs[f.focus].Focus()
	f.err = ""
}

// values snapshots the current input strings.
func (f *Form) values() formValues {
	return formValues{
		Name:          f.inputs[0].Value(),
		Host:          f.inputs[1].Value(),
		User:          f.inputs[2].Value(),
		Port:          f.inputs[3].Value(),
		IdentityFile:  f.inputs[4].Value(),
		CustomCommand: f.inputs[5].Value(),
		Group:         f.inputs[6].Value(),
		Notes:         f.inputs[7].Value(),
	}
}

// connection validates and returns the edited connection.
func (f *Form) connection() (core.Connection, error) {
	return buildConnection(f.id, f.values())
}

func (f *Form) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" "+f.title+" ") + "\n\n")
	for i, in := range f.inputs {
		label := formLabelStyle
		if i == f.focus {
			label = formLabelFocusStyle
		}
		b.WriteString(label.Render(fmt.Sprintf("%-15s", formLabels[i])))
		b.WriteString(in.View())
		b.WriteByte('\n')
	}
	if f.err != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+f.err) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("tab/↑↓ move · enter next · ctrl+s save · esc cancel"))
	return b.String()
}
