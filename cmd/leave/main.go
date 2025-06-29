package leave

import (
	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/connect"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "leave",
	Short: "close intercept, if active",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(_ *cobra.Command, args []string) error {

	cfg := config.Singleton()

	please := "please run `inkube switch` to set the app name, namespace and container"
	if cfg.Bridge.Name == "" {
		return fn.Errorf("deployment name is not set, %s", please)
	}

	if cfg.LoadEnv.Container == "" {
		return fn.Errorf("container is not set, %s", please)
	}

	if cfg.Namespace == "" {
		return fn.Errorf("namespace is not set, %s", please)
	}

	if err := connect.SClient().Leave(cfg.Bridge.Name, cfg.Namespace); err != nil {
		return err
	}

	return nil
}
