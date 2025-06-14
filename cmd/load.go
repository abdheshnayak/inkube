package cmd

import (
	"github.com/abdheshnayak/inkube/cmd/box"
	"github.com/abdheshnayak/inkube/cmd/dev"
	i "github.com/abdheshnayak/inkube/cmd/init"
	"github.com/abdheshnayak/inkube/cmd/quit"
	"github.com/abdheshnayak/inkube/cmd/status"
	sw "github.com/abdheshnayak/inkube/cmd/switch"
	"github.com/abdheshnayak/inkube/cmd/tel"
	"github.com/spf13/cobra"
)

func Load(root *cobra.Command) {
	root.AddCommand(dev.Cmd)
	root.AddCommand(i.Cmd)
	root.AddCommand(sw.Cmd)
	root.AddCommand(status.Cmd)
	root.AddCommand(quit.Cmd)
	root.AddCommand(tel.Cmd)
	root.AddCommand(box.Cmd)
}
