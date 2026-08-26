package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// execCommand replaces the current process with argv, inheriting the terminal
// and all standard streams. This preserves interactive sessions and stdin/stdout
// piping (e.g. `infocmp -x | sshm host -- tic -x -`). It only returns on error.
func execCommand(argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("empty command")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("cannot find %q: %w", argv[0], err)
	}
	return syscall.Exec(path, argv, os.Environ())
}
