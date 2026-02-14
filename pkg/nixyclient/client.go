package nixyclient

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nxtcoder17/nixy/pkg/nixy"
)

// Client ensures nixy.yml exists, nix (and optionally nixy) are present, and provides ShellEnv for the dev shell.
//
// Init uses the nixy Go library (no exec). ShellEnv execs the nixy CLI; to embed,
// build the nixy binary, embed with go:embed, extract at runtime, and use NewClientWithPath(extractedPath).
type Client struct {
	// NixyPath is the path to the nixy binary. If empty, "nixy" is used from PATH.
	// Can be set to an extracted embedded binary path for fully embedded use.
	NixyPath string
}

// DevEnvClient is the interface used by inkube for dev environment (packages + env).
type DevEnvClient interface {
	ShellEnv() (map[string]string, error)
	EnsureDependencies() error
	EnsureInit() error
}

func (c *Client) nixyBin() string {
	if c != nil && c.NixyPath != "" {
		return c.NixyPath
	}
	return "nixy"
}

func (c *Client) EnsureDependencies() error {
	// Nixy in local mode requires nix on the system.
	if _, err := exec.LookPath("nix"); err != nil {
		return fmt.Errorf("nix not found; nixy requires nix (https://nixos.org/download/)")
	}
	bin := c.nixyBin()
	if bin != "nixy" {
		if _, err := os.Stat(bin); err != nil {
			return fmt.Errorf("nixy binary not found at %s: %w", bin, err)
		}
		return nil
	}
	if _, err := exec.LookPath("nixy"); err != nil {
		return fmt.Errorf("nixy not found; install from https://github.com/nxtcoder17/nixy or set NixyPath for embedded binary")
	}
	return nil
}

func (c *Client) EnsureInit() error {
	const configFile = "nixy.yml"
	if _, err := os.Stat(configFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// File does not exist; create via nixy library (no exec).
		return nixy.InitNixyFile(context.Background(), configFile)
	}
	return nil
}

func (c *Client) ShellEnv() (map[string]string, error) {
	if err := c.EnsureInit(); err != nil {
		return nil, err
	}

	bin := c.nixyBin()
	// Nixy has no "shellenv" subcommand; we run the shell with `env` and parse (pure env).
	cmd := exec.Command(bin, "shell", "--", "bash", "-c", "env")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get nixy shell env: %w", err)
	}

	envs := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "PATH" {
				envs[parts[0]] = fmt.Sprintf("%s:%s", os.Getenv("PATH"), parts[1])
				continue
			}
			envs[parts[0]] = parts[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading nixy env output: %w", err)
	}
	return envs, nil
}

// NewClient returns a client that uses "nixy" from PATH.
func NewClient() *Client {
	return &Client{}
}

// NewClientWithPath returns a client that uses the given nixy binary (e.g. extracted from embed).
func NewClientWithPath(nixyPath string) *Client {
	return &Client{NixyPath: nixyPath}
}

// DefaultNixyConfigPath returns the default nixy config filename in the current directory.
func DefaultNixyConfigPath() string {
	return filepath.Join(".", "nixy.yml")
}
