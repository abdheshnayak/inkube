package dev

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dev",
	Short: "get dev shell, with cluster connection, ",
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

	if cfg.Tele.Connect {
		if err := fn.ExecCmd(fmt.Sprintf("telepresence connect -n %s", cfg.Namespace), nil, true); err != nil {
			return err
		}

		defer func() {
			if err := fn.ExecCmd(fmt.Sprintf("telepresence quit"), nil, true); err != nil {
				fn.PrintError(err)
			}
		}()
	}

	if cfg.Tele.Intercept {
		if err := fn.ExecCmd(fmt.Sprintf("telepresence intercept %s", cfg.Tele.Name), nil, true); err != nil {
			return err
		}

	}

	defer func() {
		if err := cfg.Reload(); err != nil {
			fn.PrintError(err)
		}
		if cfg.Tele.Intercept {
			if err := fn.ExecCmd(fmt.Sprintf("telepresence leave %s", cfg.Tele.Name), nil, true); err != nil {
				fn.PrintError(err)
			}
		}
	}()

	envs := make(map[string]string)
	if cfg.LoadEnv.Enabled {
		kubeclient := kube.Singleton()

		name := cfg.Tele.Name
		if cfg.LoadEnv.Name != nil {
			name = *cfg.LoadEnv.Name
		}

		var err error
		envs, err = kubeclient.GetEnvs(cfg.Namespace, name, cfg.LoadEnv.Container)
		if err != nil {
			return err
		}

		maps.Copy(envs, cfg.LoadEnv.Overrides)
	}

	if cfg.Devbox {
		out, err := exec.Command("devbox", "shellenv", "--pure").Output()
		if err != nil {
			return fmt.Errorf("failed to get devbox shell env: %w", err)
		}

		td := os.TempDir()
		if err := os.WriteFile(path.Join(td, "env.sh"), out, 0o644); err != nil {
			return fmt.Errorf("failed to write env.sh: %w", err)
		}

		out, err = exec.Command("bash", "-c", fmt.Sprintf(`
        env -i bash -c '
            source %s/env.sh
            env
        '
    `, td)).Output()

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				if parts[0] == "PATH" {
					envs[parts[0]] = fmt.Sprintf("%s:%s", os.Getenv("PATH"), envs[parts[0]])
					continue
				}

				envs[parts[0]] = parts[1]
			}
		}

		// cmd := exec.Command("bash", "-c", "")
		// cmd.Env = []string{}
	}

	envs["INKUBE"] = "true"
	shell, ok := os.LookupEnv("SHELL")
	if !ok {
		shell = "sh"
	}

	fn.Log(text.Blue("[#] entering inkube shell"))

	if err := fn.ExecCmd(shell, envs, true); err != nil {
		return err
	}

	fn.Log(text.Blue("[#] exited from inkube shell"))
	return nil
}
