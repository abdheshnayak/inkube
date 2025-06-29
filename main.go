package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/abdheshnayak/inkube/cmd"
	"github.com/abdheshnayak/inkube/flags"
	"github.com/abdheshnayak/inkube/pkg/connect"
	"github.com/abdheshnayak/inkube/pkg/devbox"
	fn "github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "inkube",
	Short: "Develop inside kubernetes",

	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if s, ok := os.LookupEnv("KL_DEV"); ok && s == "true" {
			flags.DevMode = "true"
		} else if ok && s == "false" {
			flags.DevMode = "false"
		}

		verbose := fn.ParseBoolFlag(cmd, "verbose")
		if verbose {
			spinner.Client.SetVerbose(verbose)
			flags.IsVerbose = verbose
		}

		quiet := fn.ParseBoolFlag(cmd, "quiet")
		if quiet {
			spinner.Client.SetQuiet(quiet)
			flags.IsQuiet = quiet
		}

		sigChan := make(chan os.Signal, 1)

		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan

			spinner.Client.Stop()
			os.Exit(1)
		}()
	},

	PersistentPostRun: func(*cobra.Command, []string) {
		spinner.Client.Stop()
	},
}

func main() {
	if err := Run(); err != nil {
		fn.PrintError(err)
		os.Exit(1)
	}
}

func Run() error {

	if err := devbox.NewDevboxClient().EnsureDependencies(); err != nil {
		return err
	}

	if err := connect.SClient().EnsureDependencies(); err != nil {
		return err
	}

	cmd.Load(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
