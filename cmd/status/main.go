package status

import (
	"os"

	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(cmd *cobra.Command, args []string) error {
	if fn.ParseBoolFlag(cmd, "simple") {
		s, ok := os.LookupEnv("INKUBE")
		if ok && s == "true" {
			fn.Logf(text.Blue("(inkube)"))
		} else {
			fn.Logf(text.Blue("(not inkube)"))
		}
		return nil
	}

	if err := fn.ExecCmd("telepresence status", nil, true); err != nil {
		return err
	}

	s, ok := os.LookupEnv("INKUBE")
	if ok && s == "true" {
		fn.Log(text.Blue("\n\nYou are in inkube session"))
	} else {
		fn.Log(text.Blue("\n\nYou are not in inkube session"))
	}
	return nil
}

func init() {
	Cmd.Flags().BoolP("simple", "s", false, "simple output")
}
