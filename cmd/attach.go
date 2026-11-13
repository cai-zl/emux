package cmd

import (
	"github.com/spf13/cobra"
)

func newAttachCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "attach [session]",
		Short: "attach to a session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 1 {
				return a.tmux.Run("attach-session", "-t", args[0])
			}
			ss, err := a.tmux.ListSessions()
			if err != nil {
				return err
			}
			out := c.OutOrStdout()
			switch {
			case len(ss) == 0:
				return sayf(out, 1, "no sessions; create one with: emux new")
			case len(ss) == 1:
				return a.tmux.Run("attach-session", "-t", ss[0].Name)
			default:
				printSessions(out, ss)
				return sayf(out, 1, "which one? run: emux attach <name>")
			}
		},
	}
}
