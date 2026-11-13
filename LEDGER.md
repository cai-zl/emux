# gmux implementation ledger

Spec: approved Plan-mode plan (gmux — a friendlier tmux wrapper).

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
- [ ] 10. Help system + version
- [ ] 11. Final verification + review

Log:
- Tasks 4–9 done in one TDD cycle (root routing + all subcommands + version; 37 tests green, go vet clean). Version = `gmux version` subcommand + bare `--version` on root.
- Ruling: tests must neutralize ambient `TMUX` (`t.Setenv("TMUX", "")`) — dev shell runs inside tmux and detach's inside/outside branch reads it. detach.go uses `os.LookupEnv` + non-empty check for robustness.
- Ruling: cobra `SetArgs(nil)` falls back to `os.Args[1:]` — test helper normalizes nil args to `[]string{}` or bare-invocation tests execute the go test binary's flags.
- Ruling: `gmux kill` with no arg lists sessions + hint and exits 0 (nothing killed ≠ failure); ambiguous `attach`/`detach` still exit 1 because they were asked to do something they couldn't.
