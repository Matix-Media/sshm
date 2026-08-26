package tui

import (
	"fmt"
	"io"

	"github.com/Matix-Media/sshm/core"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// connItem adapts a core.Connection to list.Item.
type connItem struct{ c core.Connection }

func (i connItem) FilterValue() string {
	return i.c.Name + " " + i.c.Host + " " + i.c.User + " " + i.c.Group
}

// toItems wraps connections as list items (already sorted by the caller).
func toItems(conns []core.Connection) []list.Item {
	items := make([]list.Item, len(conns))
	for i, c := range conns {
		items[i] = connItem{c}
	}
	return items
}

// userHost renders the "user@host[:port]" summary line.
func userHost(c core.Connection) string {
	s := fmt.Sprintf("%s@%s", c.User, c.Host)
	if c.Port != "" && c.Port != "22" {
		s += ":" + c.Port
	}
	return s
}

// connDelegate renders each connection as a two-line, styled entry.
type connDelegate struct{}

func (connDelegate) Height() int                         { return 2 }
func (connDelegate) Spacing() int                        { return 1 }
func (connDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (connDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(connItem)
	if !ok {
		return
	}
	c := it.c
	width := m.Width()
	selected := index == m.Index()

	bar := "  "
	name := truncate(c.Name, width-4)
	if selected {
		bar = selBarStyle.Render("▎ ")
		name = itemNameSelStyle.Render(name)
	} else {
		name = itemNameStyle.Render(name)
	}

	badge := ""
	if c.Group != "" {
		badge = "  " + badgeStyle.Render(c.Group)
	}
	// Second line: connection summary plus the id, so it's usable as the
	// `sshm <id>` reference.
	summary := truncate(userHost(c), width-len(c.ID)-8)
	detail := itemDetailStyle.Render(summary) + idStyle.Render("  "+c.ID)

	fmt.Fprintf(w, "%s%s%s\n  %s", bar, name, badge, detail)
}

// truncate shortens s to max runes, appending an ellipsis when cut.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
