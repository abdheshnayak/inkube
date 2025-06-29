// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package shell

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/adrg/xdg"
	"github.com/pkg/errors"
)

//go:embed shellrc.tmpl
var shellrcText string
var shellrcTmpl = template.Must(template.New("shellrc").Parse(shellrcText))

//go:embed shellrc_fish.tmpl
var fishrcText string
var fishrcTmpl = template.Must(template.New("shellrc_fish").Parse(fishrcText))

type name string

const (
	shUnknown name = ""
	shBash    name = "bash"
	shZsh     name = "zsh"
	shKsh     name = "ksh"
	shFish    name = "fish"
	shPosix   name = "posix"
)

var ErrNoRecognizableShellFound = errors.New("SHELL in undefined, and couldn't find any common shells in PATH")

type Inkube struct{}

// TODO consider splitting this struct's functionality so that there is a simpler
// `nix.Shell` that can produce a raw nix shell once again.

// InkubeShell configures a user's shell to run in Inkube. Its zero value is a
// fallback shell that launches a regular Nix shell.
type InkubeShell struct {
	Inkube          *Inkube
	Name            name
	BinPath         string
	ProjectDir      string // path to where inkube.json config resides
	Env             map[string]string
	UserShellrcPath string

	HistoryFile string

	// ShellStartTime is the unix timestamp for when the command was invoked
	ShellStartTime time.Time
}

type ShellOption func(*InkubeShell)

// newShell initializes the InkubeShell struct so it can be used to start a shell environment
// for the inkube project.
func (d *Inkube) NewShell(envOpts EnvOptions, opts ...ShellOption) (*InkubeShell, error) {
	shPath, err := d.shellPath(envOpts)
	if err != nil {
		return nil, err
	}
	sh := initShellBinaryFields(shPath)
	sh.Inkube = d

	for _, opt := range opts {
		opt(sh)
	}

	slog.Debug("detected user shell", "shell", sh.BinPath, "initrc", sh.UserShellrcPath)
	return sh, nil
}

// shellPath returns the path to a shell binary, or error if none found.
func (d *Inkube) shellPath(envOpts EnvOptions) (path string, err error) {
	defer func() {
		if err != nil {
			path = filepath.Clean(path)
		}
	}()

	if !envOpts.Pure {
		// First, check the SHELL environment variable.
		path = os.Getenv("SHELL")
		if path != "" {
			slog.Debug("using SHELL env var for shell binary path", "shell", path)
			return path, nil
		}
	}

	// Second, fallback to using the bash that nix uses by default.

	// Else, return an error
	return "", ErrNoRecognizableShellFound
}

// initShellBinaryFields initializes the fields specific to the shell binary that will be used
// for the inkube shell.
func initShellBinaryFields(path string) *InkubeShell {
	shell := &InkubeShell{BinPath: path}
	base := filepath.Base(path)
	// Login shell
	if base[0] == '-' {
		base = base[1:]
	}
	switch base {
	case "bash":
		shell.Name = shBash
		shell.UserShellrcPath = rcfilePath(".bashrc")
	case "zsh":
		shell.Name = shZsh
		if zdotdir := os.Getenv("ZDOTDIR"); zdotdir != "" {
			shell.UserShellrcPath = filepath.Join(os.ExpandEnv(zdotdir), ".zshrc")
		} else {
			shell.UserShellrcPath = rcfilePath(".zshrc")
		}
	case "ksh":
		shell.Name = shKsh
		shell.UserShellrcPath = rcfilePath(".kshrc")
	case "fish":
		shell.Name = shFish
		shell.UserShellrcPath = fishConfig()
	case "dash", "ash", "shell":
		shell.Name = shPosix
		shell.UserShellrcPath = os.Getenv("ENV")

		// Just make up a name if there isn't already an init file set
		// so we have somewhere to put a new one.
		if shell.UserShellrcPath == "" {
			shell.UserShellrcPath = ".shinit"
		}
	default:
		shell.Name = shUnknown
	}
	return shell
}

func WithHistoryFile(historyFile string) ShellOption {
	return func(s *InkubeShell) {
		s.HistoryFile = historyFile
	}
}

// TODO: Consider removing this once plugins add env vars directly to binaries via wrapper scripts.
func WithEnvVariables(envVariables map[string]string) ShellOption {
	return func(s *InkubeShell) {
		s.Env = envVariables
	}
}

func WithProjectDir(projectDir string) ShellOption {
	return func(s *InkubeShell) {
		s.ProjectDir = projectDir
	}
}

func WithShellStartTime(t time.Time) ShellOption {
	return func(s *InkubeShell) {
		s.ShellStartTime = t
	}
}

// rcfilePath returns the absolute path for an rcfile, which is usually in the
// user's home directory. It doesn't guarantee that the file exists.
func rcfilePath(basename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, basename)
}

func fishConfig() string {
	s, err := xdg.ConfigFile("fish/config.fish")
	if err != nil {
		return ""
	}
	return s
}

func (s *InkubeShell) Run() error {
	var cmd *exec.Cmd
	shellrc, err := s.writeInkubeShellrc()
	if err != nil {
		// We don't have a good fallback here, since all the variables we need for anything to work
		// are in the shellrc file. For now let's fail. Later on, we should remove the vars from the
		// shellrc file. That said, one of the variables we have to evaluate ($shellHook), so we need
		// the shellrc file anyway (unless we remove the hook somehow).
		slog.Error("failed to write inkube shellrc", "err", err)
		return errors.WithStack(err)
	}

	// Link other files that affect the shell settings and environments.
	s.linkShellStartupFiles(filepath.Dir(shellrc))
	extraEnv, extraArgs := s.shellRCOverrides(shellrc)
	env := s.Env
	if env == nil {
		env = make(map[string]string)
	}

	maps.Copy(env, extraEnv)

	env["SHELL"] = s.BinPath

	cmd = exec.Command(s.BinPath)
	cmd.Env = MapToPairs(env)

	cmd.Args = append(cmd.Args, extraArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Debug("Executing shell %s with args: %v", s.BinPath, cmd.Args)
	err = cmd.Run()

	// If the error is an ExitError, this means the shell started up fine but there was
	// an error from executing a shell command or script.
	//
	// This could be from one of the generated shellrc commands, but more likely is from
	// a user's command or script. So, we want to return nil for this.
	if exitErr := (&exec.ExitError{}); errors.As(err, &exitErr) {
		return nil
	}

	// This means that there was an error from inkube's code or nix's code. Not a user
	// error and so we do return it.
	return errors.WithStack(err)
}

func (s *InkubeShell) shellRCOverrides(shellrc string) (extraEnv map[string]string, extraArgs []string) {
	// Shells have different ways of overriding the shellrc, so we need to
	// look at the name to know which env vars or args to set when launching the shell.
	switch s.Name {
	case shBash:
		extraArgs = []string{"--rcfile", shellescape.Quote(shellrc)}
	case shZsh:
		extraEnv = map[string]string{"ZDOTDIR": shellescape.Quote(filepath.Dir(shellrc))}
	case shKsh, shPosix:
		extraEnv = map[string]string{"ENV": shellescape.Quote(shellrc)}
	case shFish:
		extraArgs = []string{"-C", ". " + shellrc}
	}
	return extraEnv, extraArgs
}

func (s *InkubeShell) writeInkubeShellrc() (path string, err error) {
	// We need a temp dir (as opposed to a temp file) because zsh uses
	// ZDOTDIR to point to a new directory containing the .zshrc.
	tmp, err := os.MkdirTemp("", "inkube")
	if err != nil {
		return "", fmt.Errorf("create temp dir for shell init file: %v", err)
	}

	// This is a best-effort to include the user's existing shellrc.
	userShellrc := []byte{}
	if s.UserShellrcPath != "" {
		// If we can't read it, then just omit it from the inkube shellrc.
		userShellrc, _ = os.ReadFile(s.UserShellrcPath)
	}

	// If the user already has a shellrc file, then give the inkube shellrc
	// file the same name. Otherwise, use an arbitrary name of "shellrc".
	shellrcName := "shellrc"
	if s.UserShellrcPath != "" {
		shellrcName = filepath.Base(s.UserShellrcPath)
	}
	path = filepath.Join(tmp, shellrcName)
	shellrcf, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("write to shell init file: %v", err)
	}
	defer func() {
		cerr := shellrcf.Close()
		if err == nil {
			err = cerr
		}
	}()

	tmpl := shellrcTmpl
	if s.Name == shFish {
		tmpl = fishrcTmpl
	}

	err = tmpl.Execute(shellrcf, struct {
		ProjectDir       string
		OriginalInit     string
		OriginalInitPath string
		HooksFilePath    string
		ShellStartTime   string
		HistoryFile      string
		ExportEnv        string

		RefreshAliasName   string
		RefreshCmd         string
		RefreshAliasEnvVar string
	}{
		ProjectDir:       s.ProjectDir,
		OriginalInit:     string(bytes.TrimSpace(userShellrc)),
		OriginalInitPath: s.UserShellrcPath,
		HistoryFile:      strings.TrimSpace(s.HistoryFile),
		RefreshAliasName: "inkube-refresh",
	})
	if err != nil {
		return "", fmt.Errorf("execute shellrc template: %v", err)
	}

	slog.Debug("wrote inkube shellrc", "path", path)
	return path, nil
}

// linkShellStartupFiles will link files used by the shell for initialization.
// We choose to link instead of copy so that changes made outside can be reflected
// within the inkube shell.
//
// We do not link the .{shell}rc files, since inkube modifies them. See writeInkubeShellrc
func (s *InkubeShell) linkShellStartupFiles(shellSettingsDir string) {
	// For now, we only need to do this for zsh shell
	if s.Name == shZsh {
		// List of zsh startup files: https://zsh.sourceforge.io/Intro/intro_3.html
		filenames := []string{".zshenv", ".zprofile", ".zlogin", ".zlogout"}

		// zim framework
		// https://zimfw.sh/docs/install/
		filenames = append(filenames, ".zimrc")

		for _, filename := range filenames {
			// The userShellrcPath should be set to ZDOTDIR already.
			fileOld := filepath.Join(filepath.Dir(s.UserShellrcPath), filename)
			_, err := os.Stat(fileOld)
			if errors.Is(err, fs.ErrNotExist) {
				// this file may not be relevant for the user's setup.
				continue
			}
			if err != nil {
				slog.Debug("os.Stat error for %s is %v", fileOld, err)
			}

			fileNew := filepath.Join(shellSettingsDir, filename)
			cmd := exec.Command("cp", fileOld, fileNew)
			if err := cmd.Run(); err != nil {
				// This is a best-effort operation. If there's an error then log it for visibility but continue.
				slog.Error("error copying zsh setting file", "from", fileOld, "to", fileNew, "err", err)
				continue
			}
		}
	}
}

func filterPathList(pathList string, keep func(string) bool) string {
	filtered := []string{}
	for _, path := range filepath.SplitList(pathList) {
		if keep(path) {
			filtered = append(filtered, path)
		}
	}
	return strings.Join(filtered, string(filepath.ListSeparator))
}

func isFishShell() bool {
	return filepath.Base(os.Getenv("SHELL")) == "fish" ||
		os.Getenv("FISH_VERSION") != ""
}
