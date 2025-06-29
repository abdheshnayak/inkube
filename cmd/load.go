package cmd

import (
	"github.com/abdheshnayak/inkube/cmd/connect"
	"github.com/abdheshnayak/inkube/cmd/dev"
	"github.com/abdheshnayak/inkube/cmd/disconnect"
	i "github.com/abdheshnayak/inkube/cmd/init"
	"github.com/abdheshnayak/inkube/cmd/intercept"
	"github.com/abdheshnayak/inkube/cmd/leave"
	"github.com/abdheshnayak/inkube/cmd/quit"
	"github.com/abdheshnayak/inkube/cmd/status"
	sw "github.com/abdheshnayak/inkube/cmd/switch"
	"github.com/abdheshnayak/inkube/flags"
	"github.com/spf13/cobra"
)

func Load(root *cobra.Command) {

	root.CompletionOptions.HiddenDefaultCmd = true

	root.AddCommand(dev.Cmd)
	root.AddCommand(i.Cmd)
	root.AddCommand(sw.Cmd)
	root.AddCommand(status.Cmd)
	root.AddCommand(quit.Cmd)

	root.AddCommand(intercept.Cmd)
	root.AddCommand(leave.Cmd)

	root.AddCommand(connect.Cmd)
	root.AddCommand(disconnect.Cmd)

	Init(root)
}

func Init(root *cobra.Command) {
	root.Version = flags.Version

	for _, c := range append(root.Commands(), root) {
		c.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
		c.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	}
}
