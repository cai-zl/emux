package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDetachCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "detach [session]",
		Short: "detach the current client, or a named session's clients",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 1 {
				return a.tmux.Run("detach-client", "-s", args[0])
			}
			// No name: inside tmux detach ourselves, outside fall back to
			// the single-session case instead of guessing.
			if v, ok := os.LookupEnv("TMUX"); ok && v != "" {
				return a.tmux.Run("detach-client")
			}
			ss, err := a.tmux.ListSessions()
			if err != nil {
				return err
			}
			out := c.OutOrStdout()
			switch {
			case len(ss) == 0:
				fmt.Fprintln(out, "no active sessions")
				return nil
			case len(ss) == 1:
				return a.tmux.Run("detach-client", "-s", ss[0].Name)
			default:
				printSessions(out, ss)
				return sayf(out, 1, "which one? run: easymux detach <name>")
			}
		},
	}
}
