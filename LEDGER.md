# gmux implementation ledger

Spec: approved Plan-mode plan (gmux — a friendlier tmux wrapper).

Rulings:
- Ruling: skip git worktree — empty brand-new project, nothing to isolate; `git init` in place. Cost: none (no pre-existing code or branch to disturb).
- Ruling: work directly on `main` for the same reason (no existing history to protect).

## Progress

- [ ] 1. Scaffold project
- [ ] 2. Catalog package (embedded YAML)
- [ ] 3. Tmux runner
- [ ] 4. Root routing + passthrough
- [ ] 5. ls
- [ ] 6. new
- [ ] 7. attach
- [ ] 8. kill
- [ ] 9. rename + detach
- [ ] 10. Help system + version
- [ ] 11. Final verification + review
