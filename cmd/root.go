package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "gmux",
		Short: "a friendlier tmux",
		Long:  "gmux wraps tmux: friendlier common commands, better help, everything else passes straight through to tmux.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	return root
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
