package main

import (
	"os"
	"os/exec"

	"github.com/abdheshnayak/inkube/cmd"
	fn "github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "inkube",
	Short: "Develop inside kubernetes",
}

func main() {
	if err := Run(); err != nil {
		fn.PrintError(err)
		os.Exit(1)
	}
}

func Run() error {
	_, err := exec.LookPath("devbox")
	if err != nil {
		return fn.Errorf("inkube not found, please ensure inkube is installed")
	}

	_, err = exec.LookPath("kubevpn")
	if err != nil {
		return fn.Errorf("telepresence not found, please ensure telepresence is installed")
	}

	cmd.Load(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
