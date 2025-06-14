package nixpkghandler

import (
	"context"

	"github.com/abdheshnayak/inkube/pkg/config"
	"github.com/spf13/cobra"
)

type PackageClient interface {
	AddPackage(name, hash string) error
	RemovePackage(name string) error

	AddLibrary(name, hash string) error
	RemoveLibrary(name string) error

	// used for listing search results
	Search(name string) (*SearchResults, error)

	// used for fzf search
	Find(pname string) (string, string, error)
	SyncLockfile() error

	EvaluateShell(ctx context.Context, pkgs []string, libs []string, envMap map[string]string) (map[string]string, error)
}

type pkgHandler struct {
	cmd *cobra.Command
	fc  config.ConfigClient
}

func New(cmd *cobra.Command) (PackageClient, error) {
	fc, err := config.Singleton()
	if err != nil {
		return nil, err
	}
	return &pkgHandler{
		cmd: cmd,
		fc:  fc,
	}, nil
}
