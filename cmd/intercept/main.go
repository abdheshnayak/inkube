package intercept

import (
	"fmt"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept the deployment and tunnel all traffic to the local machine",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(_ *cobra.Command, args []string) error {

	cfg := config.Singleton()

	please := "please run `inkube switch` to set the app name, namespace and container"
	if cfg.Tele.Name == "" {
		return fn.Errorf("deployment name is not set, %s", please)
	}

	if cfg.LoadEnv.Container == "" {
		return fn.Errorf("container is not set, %s", please)
	}

	if cfg.Namespace == "" {
		return fn.Errorf("namespace is not set, %s", please)
	}

	cfg.Tele.Intercept = true
	if err := cfg.Write(); err != nil {
		return err
	}

	if err := fn.ExecCmd(fmt.Sprintf("telepresence intercept %s", cfg.Tele.Name), nil, true); err != nil {
		return err
	}
	return nil
}
