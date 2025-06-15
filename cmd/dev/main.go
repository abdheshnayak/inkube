package dev

import (
	"fmt"
	"os"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dev",
	Short: "dev",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(_ *cobra.Command, args []string) error {

	cfg := config.Singleton()

	please := "please run `inkube switch` to set the app name, namespace and container"
	if cfg.Intercept.Name == "" {
		return fn.Errorf("deployment name is not set, %s", please)
	}

	if cfg.LoadEnv.Container == "" {
		return fn.Errorf("container is not set, %s", please)
	}

	if cfg.Namespace == "" {
		return fn.Errorf("namespace is not set, %s", please)
	}

	if err := fn.ExecCmd(fmt.Sprintf("telepresence connect -n %s", cfg.Namespace), nil, true); err != nil {
		return err
	}

	defer func() {
		if err := fn.ExecCmd(fmt.Sprintf("telepresence quit"), nil, true); err != nil {
			fn.PrintError(err)
		}
	}()

	if cfg.Intercept.Enabled {
		if err := fn.ExecCmd(fmt.Sprintf("telepresence intercept %s", cfg.Intercept.Name), nil, true); err != nil {
			return err
		}

		defer func() {
			if err := fn.ExecCmd(fmt.Sprintf("telepresence leave %s", cfg.Intercept.Name), nil, true); err != nil {
				fn.PrintError(err)
			}
		}()
	}

	var envs map[string]string
	if cfg.LoadEnv.Enabled {
		kubeclient := kube.Singleton()

		name := cfg.Intercept.Name
		if cfg.LoadEnv.Name != nil {
			name = *cfg.LoadEnv.Name
		}

		var err error
		envs, err = kubeclient.GetEnvs(cfg.Namespace, name, cfg.LoadEnv.Container)
		if err != nil {
			return err
		}
	}

	shell, ok := os.LookupEnv("SHELL")
	if !ok {
		shell = "sh"
	}

	if err := fn.ExecCmd(shell, envs, true); err != nil {
		return err
	}

	return fn.ExecCmd("telepresence status", nil, true)
}
