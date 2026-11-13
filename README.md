# emux

**emux wraps tmux**: friendlier common commands, better help, everything else passes straight through to tmux.

```console
$ emux ls
NAME      WINDOWS  AGE   ATTACHED
work      3        42m   yes

$ emux new work
$ emux kill work -y
killed session "work"
```

## Why

tmux is powerful but terse: `tmux ls` prints a raw format, `tmux new-session -s work -d -c /tmp/proj` is a mouthful, and `tmux kill-session` kills without asking. emux keeps tmux as the engine and adds a thin, polite layer:

- **Simple subcommands** — `emux new`, `emux ls`, `emux attach`, `emux kill`, `emux rename`, `emux detach`.
- **Readable output** — `emux ls` prints an aligned table instead of tmux's colon-separated lines.
- **Safety** — `emux kill` asks for confirmation; skip it with `-y`.
- **Everything else is tmux** — any non-emux command is forwarded verbatim, flags and exit codes included.

## Install

```sh
go install github.com/cai-zl/emux@latest
```

Requires Go 1.26+ and [tmux](https://github.com/tmux/tmux) in `$PATH`.

Or build from source:

```sh
git clone https://github.com/cai-zl/emux.git
cd emux
./build.sh   # writes size-optimized binaries for linux/darwin × amd64/arm64 into dist/
```

## Usage

| Command | Does |
|---|---|
| `emux new [name]` | create a session (`-d` to stay detached, `--dir` to set the working directory) |
| `emux ls` | list sessions as a table (name, windows, age, attached) |
| `emux attach [name]` | attach to a session; with no argument, lists sessions and exits instead of guessing |
| `emux kill [name]` | kill a session, asking `y/N` first (`-y` skips) |
| `emux rename <old> <new>` | rename a session |
| `emux detach` | detach the current client, or a named session when run outside tmux |
| `emux help <command>` | help for an emux or tmux command |
| `emux version` | print the version |

Anything that isn't an emux subcommand goes to tmux untouched:

```console
$ emux split-window -h          # same as: tmux split-window -h
$ emux -2 attach main           # even tmux's own leading flags pass through
$ emux list-commands            # full tmux command list
```

Non-zero tmux exit codes are propagated, so scripts using emux behave like scripts using tmux.

## Built-in tmux help

emux ships a static catalog of ~86 tmux commands grouped by category. `emux help split-window` shows its usage and description without leaving the terminal, and `emux` alone prints the grouped overview. The catalog is hand-maintained and may lag behind new tmux releases.

## Development

```sh
go test ./...   # 42 tests, no live tmux needed (fake tmux fixtures)
go vet ./...
./build.sh
```
