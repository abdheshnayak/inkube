package fn

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abdheshnayak/inkube/pkg/ui/text"
)

func ExecCmd(cmdString string, env map[string]string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return NewE(err, "failed to parse command")
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
	}

	cmd.Env = os.Environ()
	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	if err != nil {
		return NewE(err, "failed to execute command")
	}
	// s.Stop()
	return nil
}

func Exec(cmdString string, env map[string]string) ([]byte, error) {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, NewE(err, "failed to parse command")
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	// fn.Println(cmd.String(), cmdString)
	cmd.Stderr = os.Stderr

	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return out, err
}

func ExecNoOutput(cmdString string) error {
	cmd := exec.Command("sh", "-c", cmdString)
	_, err := cmd.Output()
	if err != nil {
		return NewE(err, "failed to execute command")
	}
	return nil
}

// isAdmin checks if the current user has administrative privileges.
func isAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

func WinSudoExec(cmdString string, env map[string]string) ([]byte, error) {
	if isAdmin() {
		return Exec(cmdString, env)
	}

	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, NewE(err)
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	quotedArgs := strings.Join(cmdArr[1:], ",")

	return Exec(fmt.Sprintf("powershell -Command Start-Process -WindowStyle Hidden -FilePath %s -ArgumentList %q -Verb RunAs", cmd.Path, quotedArgs), map[string]string{"PATH": os.Getenv("PATH")})
}

func StreamOutput(ctx context.Context, cmdString string, env map[string]string, writer io.Writer, errCh chan<- error) error {
	defer close(errCh)
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, cmdArr[0], cmdArr[1:]...)

	cmd.Env = os.Environ()
	cmd.Stderr = writer
	cmd.Stdout = writer

	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func GetUidNGid() (int, int, error) {

	uidstr, ok := os.LookupEnv("SUDO_UID")
	if !ok {
		return 0, 0, Error("failed to get sudo uid")
	}

	gidstr, ok := os.LookupEnv("SUDO_GID")
	if !ok {
		return 0, 0, Error("failed to get sudo gid")
	}

	uid, err := strconv.Atoi(uidstr)
	if err != nil {
		return 0, 0, NewE(err, "failed to get sudo uid")
	}

	gid, err := strconv.Atoi(gidstr)
	if err != nil {
		return 0, 0, NewE(err, "failed to get sudo gid")
	}

	return uid, gid, nil
}

func EnvMapToSlice(env map[string]string) []string {
	var result []string
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func EnvSliceToMap(env []string) map[string]string {
	result := make(map[string]string, len(env))
	for _, kv := range env {
		key, val, found := strings.Cut(kv, "=")
		if !found {
			return nil
		}
		result[key] = val
	}
	return result
}

func WarnReload() {
	if _, ok := os.LookupEnv("KL_SHELL"); ok {
		Warn(text.Yellow("environment variables are updated, please run `reload` to reflect changes to your current shell"))
	}
}
