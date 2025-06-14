package tel

import (
	"strings"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:                "tel",
	Short:              "tel",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func run(_ *cobra.Command, args []string) error {
	args = append([]string{"telepresence"}, args...)
	return fn.ExecCmd(strings.Join(args, " "), nil, true)
}
