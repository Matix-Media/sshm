package tui

import (
	"fmt"

	"github.com/Matix-Media/sshm/core"
)

type confirmKind int

const (
	confirmDelete confirmKind = iota
	confirmImport
)

// Confirm is a small yes/no modal for destructive or scanning actions.
type Confirm struct {
	kind confirmKind
	conn core.Connection
}

func newDeleteConfirm(c core.Connection) *Confirm {
	return &Confirm{kind: confirmDelete, conn: c}
}

func newImportConfirm() *Confirm {
	return &Confirm{kind: confirmImport}
}

func (c *Confirm) View() string {
	var msg string
	switch c.kind {
	case confirmDelete:
		msg = fmt.Sprintf("Delete connection %s?", itemNameSelStyle.Render(c.conn.Name))
	case confirmImport:
		msg = "Scan ~/.ssh/config and import new connections?"
	}
	body := msg + "\n\n" + helpStyle.Render("y confirm · any other key cancel")
	return boxStyle.Render(body)
}
