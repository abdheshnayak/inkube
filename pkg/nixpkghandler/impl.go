package nixpkghandler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/abdheshnayak/inkube/flags"
	fn "github.com/abdheshnayak/inkube/pkg/functions"
	"github.com/abdheshnayak/inkube/pkg/ui/fzf"
)

type HashData map[string]string

func (p *pkgHandler) Search(query string) (*SearchResults, error) {
	return search(p.cmd.Context(), query)
}

func (p *pkgHandler) AddLibrary(name, pkghash string) error {
	if err := p.syncPackage(pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addLibToLock(name, pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addLibToKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) RemoveLibrary(name string) error {
	if err := p.rmLibFromKlFile(name); err != nil {
		return fn.NewE(err)
	}

	return p.SyncLockfile()
}

func (p *pkgHandler) AddPackage(name, pkghash string) error {
	if err := p.syncPackage(pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addPackageToLock(name, pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addPackageKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) RemovePackage(name string) error {
	if err := p.rmPackageFromKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) Find(pname string) (string, string, error) {
	var name string
	var ver string

	if !strings.Contains(pname, "@") {
		sr, err := p.Search(pname)
		if err != nil {
			return "", "", fn.NewE(err)
		}

		pkg, err := fzf.FindOne(sr.Packages, func(item Package) string {
			return item.Name
		}, fzf.WithPrompt("select a package"))

		if err != nil {
			return "", "", fn.NewE(err)
		}

		version, err := fzf.FindOne(pkg.Versions, func(item PackageVersion) string {
			return fmt.Sprintf("%s %s", item.Version, item.Summary)
		}, fzf.WithPrompt("select a version"))

		if err != nil {
			return "", "", fn.NewE(err)
		}
		name = version.Name
		ver = version.Version
	} else {
		splits := strings.Split(name, "@")

		if strings.TrimSpace(splits[0]) == "" || strings.TrimSpace(splits[1]) == "" {
			return "", "", fn.Errorf("package %s is invalid", name)
		}
		name = splits[0]
		ver = splits[1]
	}

	return p.resolve(fmt.Sprintf("%s@%s", name, ver))
}

func (p *pkgHandler) SyncLockfile() error {
	type System struct {
		AttrPaths []string `json:"attr_paths"`
	}

	type Res struct {
		CommitHash string            `json:"commit_hash"`
		Version    string            `json:"version"`
		Systems    map[string]System `json:"systems"`
	}

	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	lfcheck, err := kf.GetChecksum()

	newLock := HashData{}
	lf, err := p.fc.GetLockfile()
	if err != nil {
		return fn.NewE(err)
	}

	if lf.Checksum == lfcheck {
		return nil
	}

	for _, v := range kf.Packages {
		if hash, ok := lf.Packages[v]; ok {
			newLock[v] = hash
			continue
		}

		_, pkgHash, err := p.resolve(v)
		if err != nil {
			return fn.NewE(err)
		}

		newLock[v] = pkgHash
	}
	lf.Packages = newLock

	newLock = HashData{}
	for _, v := range kf.Libraries {
		if hash, ok := lf.Libraries[v]; ok {
			newLock[v] = hash
			continue
		}

		_, pkgHash, err := p.resolve(v)
		if err != nil {
			return fn.NewE(err)
		}

		newLock[v] = pkgHash
	}

	lf.Libraries = newLock
	lf.Checksum = lfcheck

	fn.WarnReload()
	return lf.Save()
}

func (p *pkgHandler) EvaluateShell(ctx context.Context, packages []string, libraries []string, envMap map[string]string) (map[string]string, error) {
	resp := make(map[string]string)

	path, err := installPackage(packages...)
	if err != nil {
		return nil, fn.NewE(err)
	}

	resp["KL_NIX_PATH"] = strings.TrimSpace(path)

	libPaths := make([]string, 0, len(libraries))
	var includes []string

	for _, lib := range libraries {
		c := exec.CommandContext(ctx, "nix", "eval", lib, "--raw")
		c.Env = fn.EnvMapToSlice(envMap)
		if flags.IsVerbose {
			fn.Log(c.String())
		}

		b, err := c.CombinedOutput()
		if err != nil {
			return nil, fn.NewE(err)
		}

		if pathExists(string(b) + "/lib") {
			libPaths = append(libPaths, string(b)+"/lib")
		}

		if pathExists(string(b) + "/include") {
			includes = append(includes, string(b)+"/include")
		}

		cmd := exec.CommandContext(ctx, "nix-store", "--query", "--references", string(b))
		cmd.Env = fn.EnvMapToSlice(envMap)

		if flags.IsVerbose {
			fn.Log(cmd.String())
		}

		coutput, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fn.NewE(err)
		}
		lines := strings.Split(string(coutput), "\n")

		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 && !strings.Contains(line, "-glibc-") {
				if pathExists(line + "/lib") {
					libPaths = append(libPaths, line+"/lib")
				}
			}
		}
	}

	libPaths = createSet(libPaths)
	includes = createSet(includes)

	resp["LD_LIBRARY_PATH"] = strings.ReplaceAll(fmt.Sprintf("%s:%s", strings.Join(libPaths, ":"), os.Getenv("LD_LIBRARY_PATH")), "::", ":")
	resp["CPATH"] = strings.ReplaceAll(fmt.Sprintf("%s:%s", strings.Join(includes, ":"), os.Getenv("CPATH")), "::", ":")

	return resp, nil
}

func installPackage(pkgs ...string) (string, error) {
	penvPath, err := exec.LookPath("printenv")
	if err != nil {
		return "", err
	}

	nixPath, err := exec.LookPath("nix")
	if err != nil {
		return "", err
	}

	c := exec.Command("sh", "-c", fmt.Sprintf("%s shell %s --command printenv PATH", nixPath, strings.Join(pkgs, " ")))
	c.Env = []string{fmt.Sprintf("PATH=%s", path.Dir(penvPath))}

	if flags.IsVerbose {
		fn.Log(c.String())
	}

	b := new(bytes.Buffer)
	c.Stdout = b
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return "", err
	}

	return b.String(), nil
}
