package dev

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/abdheshnayak/inkube/pkg/connect"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/kube"
	"github.com/abdheshnayak/inkube/pkg/shell"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dev",
	Short: "start devlopment shell, you will get cluster connection, packages, and env vars of deployed app",
	Run: func(cmd *cobra.Command, args []string) {

		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(cmd *cobra.Command, args []string) error {

	tele := connect.SClient()
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

	if cfg.Connect {
		if err := tele.Connect(cfg.Namespace); err != nil {
			return err
		}

		defer func() {
			if err := tele.Disconnect(); err != nil {
				fn.PrintError(err)
			}
		}()
	}

	// defer func() {
	// 	if err := cfg.Reload(); err != nil {
	// 		fn.PrintError(err)
	// 	}
	// }()

	envs := make(map[string]string)
	if cfg.LoadEnv.Enabled {
		kubeclient := kube.Singleton()

		name := cfg.Bridge.Name
		if cfg.LoadEnv.Name != nil {
			name = *cfg.LoadEnv.Name
		}

		var err error

		refetch := fn.ParseBoolFlag(cmd, "refetch")
		envs, err = kubeclient.GetEnvs(cfg.Namespace, name, cfg.LoadEnv.Container, refetch)
		if err != nil {
			return err
		}

		maps.Copy(envs, cfg.LoadEnv.Overrides)
	}

	if cfg.Devbox {
		out, err := exec.Command("inkube", "shellenv", "--pure").Output()
		if err != nil {
			return fmt.Errorf("failed to get inkube shell env: %w", err)
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
	}

	envs["INKUBE"] = "true"

	fn.Log(text.Blue("[#] entering inkube shell"))

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	envMaps := shell.PairsToMap(os.Environ())
	maps.Copy(envMaps, envs)

	ds, err := (&shell.Inkube{}).NewShell(shell.EnvOptions{}, shell.WithProjectDir(dir), shell.WithShellStartTime(time.Now()), shell.WithEnvVariables(envMaps))

	if err := ds.Run(); err != nil {
		return err
	}

	fn.Log(text.Blue("[#] exited from inkube shell"))
	return nil
}

func init() {
	Cmd.Flags().BoolP("refetch", "r", false, "refetch env vars from cluster")
}
