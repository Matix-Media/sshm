// Command sshm is an SSH connection manager. Run with no arguments for the
// interactive TUI; pass a connection id or name to open it directly like the
// ssh CLI, forwarding any trailing arguments (so pipes and remote commands work,
// e.g. `infocmp -x | sshm host -- tic -x -`).
package main

import (
	"fmt"
	"os"

	"github.com/Matix-Media/sshm/core"
	"github.com/Matix-Media/sshm/tui"
)

const helpText = `sshm — SSH connection manager

Usage:
  sshm                        Open the interactive connection manager
  sshm list, sshm ls          List saved connections
  sshm <id|name> [args...]    Connect directly, forwarding args to ssh
  sshm -h, --help             Show this help

Examples:
  sshm list                   Show all saved connections
  sshm colima                 Connect to the "colima" profile
  sshm web -- uptime          Run "uptime" on the "web" profile
  infocmp -x | sshm web -- tic -x -
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		runInteractive()
		return
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(helpText)
	case "list", "ls":
		runList()
	default:
		runDirect(args[0], args[1:])
	}
}

// runInteractive shows the TUI and, if the user selects a connection, hands off
// to ssh after the program has torn down the alt-screen.
func runInteractive() {
	if empty, err := autoImportIfEmpty(); err != nil {
		fail(err)
	} else if empty {
		// nothing yet; the TUI will still open so the user can add/import.
	}

	conn, err := tui.Run()
	if err != nil {
		fail(err)
	}
	if conn != nil {
		connect(*conn, nil)
	}
}

// runList prints the saved connections as a plain-text table.
func runList() {
	conns, err := core.Load()
	if err != nil {
		fail(err)
	}
	fmt.Print(core.FormatList(conns))
}

// runDirect resolves a reference and execs ssh immediately, keeping stdout clean
// for piping. Errors go to stderr.
func runDirect(ref string, extra []string) {
	conns, err := core.Load()
	if err != nil {
		fail(err)
	}
	conn, err := core.Resolve(conns, ref)
	if err != nil {
		fail(err)
	}
	connect(conn, extra)
}

// connect builds the command and replaces this process with it.
func connect(conn core.Connection, extra []string) {
	if err := execCommand(core.BuildCommand(conn, extra)); err != nil {
		fail(err)
	}
}

// autoImportIfEmpty imports from ~/.ssh/config on first run (empty database),
// mirroring the original tool. It reports whether the store was empty.
func autoImportIfEmpty() (bool, error) {
	conns, err := core.Load()
	if err != nil {
		return false, err
	}
	if len(conns) > 0 {
		return false, nil
	}
	if _, err := core.ImportSSHConfig(); err != nil {
		return true, err
	}
	return true, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "sshm:", err)
	os.Exit(1)
}
