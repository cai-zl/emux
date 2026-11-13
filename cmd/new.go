package cmd

import (
	"github.com/spf13/cobra"
)

func newNewCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [name]",
		Aliases: []string{"n"},
		Short: "create a new session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			tmuxArgs := []string{"new-session"}
			if det, _ := c.Flags().GetBool("detach"); det {
				tmuxArgs = append(tmuxArgs, "-d")
			}
			if dir, _ := c.Flags().GetString("dir"); dir != "" {
				tmuxArgs = append(tmuxArgs, "-c", dir)
			}
			if len(args) == 1 {
				tmuxArgs = append(tmuxArgs, "-s", args[0])
			}
			return a.tmux.Run(tmuxArgs...)
		},
	}
	cmd.Flags().BoolP("detach", "d", false, "do not attach to the new session")
	cmd.Flags().String("dir", "", "working directory for the new session")
	return cmd
}
