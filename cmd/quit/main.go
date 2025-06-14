package quit

import (
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "quit",
	Short: "quit",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(_ *cobra.Command, args []string) error {
	return fn.ExecCmd("telepresence quit", nil, true)
}
