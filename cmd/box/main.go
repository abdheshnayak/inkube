package box

import (
	"strings"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:                "box",
	Short:              "box",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func run(_ *cobra.Command, args []string) error {
	args = append([]string{"devbox"}, args...)
	return fn.ExecCmd(strings.Join(args, " "), nil, true)
}
