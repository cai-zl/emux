package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "emux version %s\n", version)
}

func newVersionCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the emux version",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			printVersion(c.OutOrStdout())
			return nil
		},
	}
}
