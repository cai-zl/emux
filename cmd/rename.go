package cmd

import (
	"github.com/spf13/cobra"
)

func newRenameCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:    "rename <old> <new>",
		Short:  "rename a session",
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error { return a.tmux.Run("rename-session", "-t", args[0], args[1]) },
	}
}
