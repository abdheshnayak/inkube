package devbox

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/abdheshnayak/inkube/pkg/fn"
)

type DevboxClient interface {
	ShellEnv() (map[string]string, error)
	EnsureDependencies() error
	EnsureInit() error
}

type devboxClient struct {
}

func (d *devboxClient) EnsureDependencies() error {
	_, err := exec.LookPath("devbox")
	if err != nil {
		return fmt.Errorf("devbox not found, please ensure devbox is installed")
	}
	return nil
}

func (d *devboxClient) EnsureInit() error {
	if _, err := os.Stat("devbox.json"); err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) {
		return fn.ExecCmd("devbox init", nil, false)
	}
	return nil
}

func (d *devboxClient) ShellEnv() (map[string]string, error) {
	if err := d.EnsureInit(); err != nil {
		return nil, err
	}

	envs := make(map[string]string)
	out, err := exec.Command("devbox", "shellenv", "--pure").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get inkube shell env: %w", err)
	}

	td := os.TempDir()
	if err := os.WriteFile(path.Join(td, "env.sh"), out, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write env.sh: %w", err)
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

	return envs, nil
}

func NewDevboxClient() DevboxClient {
	return &devboxClient{}
}
