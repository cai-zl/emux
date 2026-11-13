# emux implementation ledger

Spec: approved Plan-mode plan (emux — a friendlier tmux wrapper).

Rulings:

- Ruling: skip git worktree — empty brand-new project, nothing to isolate; `git init` in place. Cost: none (no pre-existing code or branch to disturb).
- Ruling: work directly on `main` for the same reason (no existing history to protect).

## Progress

- [x] 1. Scaffold project
- [x] 2. Catalog package (embedded YAML)
- [x] 3. Tmux runner
- [x] 4. Root routing + passthrough
- [x] 5. ls
- [x] 6. new
- [x] 7. attach
- [x] 8. kill
- [x] 9. rename + detach
- [x] 10. Help system + version
- [x] 11. Final verification + review

Log:

- Tasks 4–9 done in one TDD cycle (root routing + all subcommands + version; 37 tests green, go vet clean). Version = `emux version` subcommand + bare `--version` on root.
- Ruling: tests must neutralize ambient `TMUX` (`t.Setenv("TMUX", "")`) — dev shell runs inside tmux and detach's inside/outside branch reads it. detach.go uses `os.LookupEnv` + non-empty check for robustness.
- Ruling: cobra `SetArgs(nil)` falls back to `os.Args[1:]` — test helper normalizes nil args to `[]string{}` or bare-invocation tests execute the go test binary's flags.
- Ruling: `emux kill` with no arg lists sessions + hint and exits 0 (nothing killed ≠ failure); ambiguous `attach`/`detach` still exit 1 because they were asked to do something they couldn't.
- Task 10 done: custom help command replaces cobra's default — emux subcommands get cobra help, tmux commands resolve from the catalog (`emux help <cmd>` → usage + description), unknown topics pass through as `tmux help <topic>`. Root `--help`/bare help embeds the grouped catalog (7 categories, 86 commands). Cobra's default completion command disabled (out of scope per plan). Version shipped earlier in the tasks 4–9 commit.
- Task 11 smoke (real tmux, live server): `emux version` → 0.1.0; `emux ls` table on real server; `new emux-smoke -d` → visible in ls → `kill emux-smoke -y` → gone; `help split-window` catalog hit; `list-commands` passthrough; `help bogus-cmd-xyz` exits 1 (tmux's own error propagated). Build: go build clean.
- Final review verdict: **Ship** (fresh reviewer, full source + cobra v1.10.2 source-traced). Routing provably clean (`stripFlags` analysis: `emux -2 attach` lands on root raw args). Applied follow-ups: RunError.Stderr now printed in Execute (silent listing failures fixed, testable `execute(root)` seam added); kill prompt moved to stderr; kill `-y` prefix-matching documented in Long; dead `var _ io.Writer` deleted. 42 tests green after fixes.
- Ruling: review follow-ups applied without new tests only where unobservable or plumbing-level (prompt stream identical in test capture; stderr print verified via the new execute seam test).
- Incident note: the reviewer delegate ran `tmux kill-server` while probing real tmux behavior, killing the live dev server (its own test sessions died with it). No code impact; disclosed to user.

## Rename: gmux → emux (user request)
- Reason: "gmux" name collided with ≥5 existing projects (incl. a same-idea Go tmux wrapper); user chose emux (visible tmux lineage via "mux"; collisions explicitly not a factor).
- Mechanical rename across 13 files (go.mod module, all imports, CLI strings, help text, tests, LEDGER). Directory not renamed (cwd in use).
- Verified: no "gmux" refs left, go vet clean, 42/42 tests pass, build OK, smoke: `emux --version`, root help, `emux ls`.

## Rename: easymux → emux (user request: shorter name)
- Same mechanical rename, 4-char name keeps the "mux" tmux lineage. Verified: vet clean, 42/42 tests, build + smoke OK.
