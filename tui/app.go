// Package tui implements sshm's interactive connection manager as a Bubble Tea
// program: a styled list plus add/edit forms and confirm modals. Selecting a
// connection quits the program and hands the chosen profile back to the caller,
// which performs the actual ssh handoff.
package tui

import (
	"fmt"

	"github.com/Matix-Media/sshm/core"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeList mode = iota
	modeForm
	modeConfirm
)

// App is the root Bubble Tea model.
type App struct {
	list    list.Model
	form    *Form
	confirm *Confirm
	mode    mode
	conns   []core.Connection
	status  string
	width   int
	height  int
	chosen  *core.Connection // set when the user picks a connection to open
}

// Run launches the manager and returns the connection the user chose to open,
// or nil if they quit without selecting one.
func Run() (*core.Connection, error) {
	conns, err := core.Load()
	if err != nil {
		return nil, err
	}
	final, err := tea.NewProgram(newApp(conns), tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}
	return final.(App).chosen, nil
}

func newApp(conns []core.Connection) App {
	core.Sort(conns)
	l := list.New(toItems(conns), connDelegate{}, 0, 0)
	l.Title = "⚡ sshm"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.AdditionalShortHelpKeys = listExtraKeys
	l.AdditionalFullHelpKeys = listExtraKeys
	return App{list: l, conns: conns, mode: modeList}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		a.width, a.height = ws.Width, ws.Height
		fw, fh := docStyle.GetFrameSize()
		a.list.SetSize(ws.Width-fw, ws.Height-fh)
		if a.form != nil {
			a.form.setWidth(ws.Width - fw)
		}
		return a, nil
	}
	switch a.mode {
	case modeForm:
		return a.updateForm(msg)
	case modeConfirm:
		return a.updateConfirm(msg)
	default:
		return a.updateList(msg)
	}
}

func (a App) View() string {
	switch a.mode {
	case modeForm:
		return docStyle.Render(a.form.View())
	case modeConfirm:
		return docStyle.Render(a.confirm.View())
	default:
		body := a.list.View()
		if a.status != "" {
			body += "\n" + statusStyle.Render(a.status)
		}
		return docStyle.Render(body)
	}
}

// updateList handles the connection list, intercepting app-level keys before
// they reach the list (except while the user is typing a filter).
func (a App) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && a.list.FilterState() != list.Filtering {
		switch {
		case key.Matches(km, keys.Quit):
			return a, tea.Quit
		case key.Matches(km, keys.Connect):
			if c, ok := a.selected(); ok {
				a.chosen = &c
				return a, tea.Quit
			}
		case key.Matches(km, keys.Add):
			return a.openForm(newAddForm(a.width))
		case key.Matches(km, keys.Edit):
			if c, ok := a.selected(); ok {
				return a.openForm(newEditForm(c, a.width))
			}
		case key.Matches(km, keys.Delete):
			if c, ok := a.selected(); ok {
				a.confirm = newDeleteConfirm(c)
				a.mode = modeConfirm
			}
			return a, nil
		case key.Matches(km, keys.Import):
			a.confirm = newImportConfirm()
			a.mode = modeConfirm
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	return a, cmd
}

func (a App) openForm(f *Form) (tea.Model, tea.Cmd) {
	a.form = f
	a.mode = modeForm
	a.status = ""
	return a, f.focusCmd()
}

func (a App) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	action, cmd := a.form.Update(msg)
	switch action {
	case formSubmit:
		conn, err := a.form.connection()
		if err != nil {
			a.form.err = err.Error()
			return a, nil
		}
		if err := a.upsert(conn); err != nil {
			a.form.err = err.Error()
			return a, nil
		}
		a.mode = modeList
		a.status = "Saved " + conn.Name
		return a, nil
	case formCancel:
		a.mode = modeList
		return a, nil
	}
	return a, cmd
}

func (a App) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}
	confirmed := key.Matches(km, keys.Yes)
	if confirmed {
		switch a.confirm.kind {
		case confirmDelete:
			if err := a.deleteByID(a.confirm.conn.ID); err != nil {
				a.status = "Delete failed: " + err.Error()
			} else {
				a.status = "Deleted " + a.confirm.conn.Name
			}
		case confirmImport:
			a.runImport()
		}
	}
	a.mode = modeList
	return a, nil
}

// selected returns the highlighted connection, if any.
func (a App) selected() (core.Connection, bool) {
	it, ok := a.list.SelectedItem().(connItem)
	if !ok {
		return core.Connection{}, false
	}
	return it.c, true
}

// upsert saves a new or edited connection and refreshes the list.
func (a *App) upsert(conn core.Connection) error {
	replaced := false
	for i := range a.conns {
		if a.conns[i].ID == conn.ID {
			a.conns[i] = conn
			replaced = true
			break
		}
	}
	if !replaced {
		a.conns = append(a.conns, conn)
	}
	return a.persist()
}

func (a *App) deleteByID(id string) error {
	out := a.conns[:0]
	for _, c := range a.conns {
		if c.ID != id {
			out = append(out, c)
		}
	}
	a.conns = out
	return a.persist()
}

func (a *App) runImport() {
	n, err := core.ImportSSHConfig()
	if err != nil {
		a.status = "Import failed: " + err.Error()
		return
	}
	if conns, err := core.Load(); err == nil {
		a.conns = conns
	}
	core.Sort(a.conns)
	a.list.SetItems(toItems(a.conns))
	a.status = fmt.Sprintf("Imported %d new connection(s)", n)
}

// persist writes the store and refreshes the list items.
func (a *App) persist() error {
	if err := core.Save(a.conns); err != nil {
		return err
	}
	core.Sort(a.conns)
	a.list.SetItems(toItems(a.conns))
	return nil
}
