package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds every binding the app reacts to. The list view reuses its own
// built-in navigation/filter keys; these are the app-level additions plus the
// form/confirm controls.
type keyMap struct {
	Connect key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Import  key.Binding
	Quit    key.Binding

	Yes    key.Binding
	Next   key.Binding
	Prev   key.Binding
	Save   key.Binding
	Cancel key.Binding
}

var keys = keyMap{
	Connect: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "connect")),
	Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Import:  key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "import")),
	Quit:    key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),

	Yes:    key.NewBinding(key.WithKeys("y", "Y")),
	Next:   key.NewBinding(key.WithKeys("tab", "down")),
	Prev:   key.NewBinding(key.WithKeys("shift+tab", "up")),
	Save:   key.NewBinding(key.WithKeys("ctrl+s")),
	Cancel: key.NewBinding(key.WithKeys("esc")),
}

// listExtraKeys are surfaced in the list's help footer.
func listExtraKeys() []key.Binding {
	return []key.Binding{keys.Connect, keys.Add, keys.Edit, keys.Delete, keys.Import}
}
