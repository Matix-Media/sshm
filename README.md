# sshm — SSH connection manager

A small, fast SSH connection manager written in Go. Save connection profiles,
browse them in a modern terminal UI, and — unlike a plain manager — open any of
them **directly from the command line like `ssh` itself**, including pipes and
remote commands.

It stores profiles in `~/.config/sshm/connections.json` and can import existing
hosts from your `~/.ssh/config`.

## Features

- **Interactive TUI** (Bubble Tea): styled list with group badges, fuzzy filter
  (`/`), and forms for adding/editing connections.
- **Use it like `ssh`**: `sshm <id|name> [args…]` execs `ssh` directly,
  forwarding any trailing arguments, so stdin/stdout piping works:
  ```sh
  infocmp -x | sshm web -- tic -x -
  ```
- **List command**: `sshm list` prints a plain-text, pipe-friendly table.
- **Import** hosts from `~/.ssh/config` (handles `Include`, multi-host lines,
  and skips wildcard patterns).
- **Custom commands** per profile (e.g. `mosh`, or an `ssh` with extra flags).

## Install

Requires Go 1.24+.

```sh
go install github.com/Matix-Media/sshm@latest
```

This installs the `sshm` binary into `$(go env GOPATH)/bin` (usually
`~/go/bin`); make sure that directory is on your `PATH`.

Or build from a local clone:

```sh
git clone https://github.com/Matix-Media/sshm.git
cd sshm
go build -o sshm .
```

## Usage

```
sshm                        Open the interactive connection manager
sshm list, sshm ls          List saved connections
sshm <id|name> [args...]    Connect directly, forwarding args to ssh
sshm -h, --help             Show help
```

Examples:

```sh
sshm                         # browse and connect in the TUI
sshm list                    # show all saved connections
sshm colima                  # connect by name (or by id: sshm 0cce1baf)
sshm web -- uptime           # run a remote command
infocmp -x | sshm web -- tic -x -   # pipe stdin through to the remote
```

Connections can be referenced by their 8-character **id** or their **name**
(case-insensitive); the id is shown in both `sshm list` and the TUI.

### Interactive keys

| Key | Action |
| --- | --- |
| `↑`/`↓`, `j`/`k` | Move selection |
| `enter` | Connect |
| `a` | Add connection |
| `e` | Edit selected |
| `d` | Delete selected |
| `i` | Import from `~/.ssh/config` |
| `/` | Filter |
| `q` | Quit |

## Configuration

Profiles live in `~/.config/sshm/connections.json`. Set `SSHM_CONFIG_DIR` to use
a different directory.

## Development

```sh
go test ./...     # run the unit tests
go vet ./...
```

Layout:

- `core/` — data model, JSON store, `ssh_config` import, id/name resolution,
  command building (pure, unit-tested).
- `tui/` — the Bubble Tea interface (list, forms, confirm modals, theme).
- `main.go` — CLI dispatch; `ssh.go` — the `syscall.Exec` handoff.
