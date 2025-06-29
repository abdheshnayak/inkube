package status

import (
	"github.com/abdheshnayak/inkube/pkg/connect"
	"github.com/abdheshnayak/inkube/pkg/fn"
	"github.com/abdheshnayak/inkube/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "get status of inkube session",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Run(cmd *cobra.Command, args []string) error {
	connected, bool, err := connect.SClient().Status()

	if fn.ParseBoolFlag(cmd, "prompt") {
		connectedStr := "✅"
		interceptedStr := "🕵️➡️💻"
		if !connected {
			connectedStr = "❌"
		}

		if !bool {
			interceptedStr = ""
		}
		fn.Printf(text.Blue("%s(inkube)%s"), connectedStr, interceptedStr)
		return nil
	}
	if err != nil {
		return err
	}

	if !connected {
		fn.Log(text.Blue("\n\nYou are not in inkube session"))
		return nil
	}

	fn.Logf(text.Blue("You are connected to cluster"))
	return nil
}

func init() {
	Cmd.Flags().BoolP("prompt", "p", false, "output for prompt")
}
