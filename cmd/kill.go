package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"emux/internal/tmux"
)

func newKillCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill [session]",
		Short: "kill a session (asks for confirmation)",
		Long:  "Kill a tmux session after a [y/N] confirmation. The session name must match exactly; with -y, tmux's own (prefix) target matching applies instead.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			out := c.ErrOrStderr()
			if len(args) == 0 {
				ss, err := a.tmux.ListSessions()
				if err != nil {
					return err
				}
				if len(ss) == 0 {
					fmt.Fprintln(out, "no active sessions")
					return nil
				}
				printSessions(out, ss)
				fmt.Fprintln(out, "which one? run: emux kill <name>")
				return nil
			}
			name := args[0]
			yes, _ := c.Flags().GetBool("yes")
			if !yes {
				ss, err := a.tmux.ListSessions()
				if err != nil {
					return err
				}
				var s *tmux.Session
				for i := range ss {
					if ss[i].Name == name {
						s = &ss[i]
						break
					}
				}
				if s == nil {
					return sayf(out, 1, "no such session: %s", name)
				}
				att := ""
				if s.Attached {
					att = " (attached)"
				}
				fmt.Fprintf(out, "kill session %q (%d windows%s)? [y/N] ", name, s.Windows, att)
				if !confirm(a.stdin()) {
					return sayf(out, 1, "aborted")
				}
			}
			if err := a.tmux.Run("kill-session", "-t", name); err != nil {
				return err
			}
			return sayf(out, 0, "killed session %q", name)
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt")
	return cmd
}

func confirm(r io.Reader) bool {
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}
