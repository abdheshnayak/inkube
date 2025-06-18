package quit

import (
	"fmt"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "quit",
	Short: "close running telepresence session, and quit intercept if active",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(_ *cobra.Command, args []string) error {
	cfg := config.Singleton()
	if cfg.Tele.Intercept {
		if err := fn.ExecCmd(fmt.Sprintf("telepresence leave %s", cfg.Tele.Name), nil, true); err != nil {
			fn.PrintError(err)
		}
	}

	return fn.ExecCmd("telepresence quit", nil, true)
}
