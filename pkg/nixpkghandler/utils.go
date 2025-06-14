package nixpkghandler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	fn "github.com/abdheshnayak/inkube/pkg/functions"
	"github.com/abdheshnayak/inkube/pkg/ui/spinner"
)

func (p *pkgHandler) resolve(pname string) (string, string, error) {
	if !strings.Contains(pname, "@") {
		pname = fmt.Sprintf("%s@latest", pname)
	}

	splits := strings.Split(pname, "@")
	if len(splits) < 2 {
		return "", "", fn.Errorf("package %s is invalid", pname)
	}

	if strings.TrimSpace(splits[0]) == "" || strings.TrimSpace(splits[1]) == "" {
		return "", "", fn.Errorf("package %s is invalid", pname)
	}

	name := splits[0]
	version := splits[1]

	type System struct {
		AttrPaths []string `json:"attr_paths"`
	}

	type Res struct {
		CommitHash string            `json:"commit_hash"`
		Version    string            `json:"version"`
		Systems    map[string]System `json:"systems"`
	}

	platform := runtime.GOARCH + runtime.GOOS
	switch runtime.GOARCH {
	case "x86_64", "amd64":
		platform = "x86_64-" + runtime.GOOS
	case "arm64":
		platform = "aarch64-" + runtime.GOOS

	}

	sr, err := caller[Res](p.cmd.Context(), fmt.Sprintf("%s/v1/resolve?name=%s&version=%s&platform=%s", searchAPIEndpoint, name, version, platform))

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", fn.Errorf("package %s not found", name)
		}
		return "", "", fn.NewE(err)
	}

	return fmt.Sprintf("%s@%s", name, sr.Version), fmt.Sprintf("%s#%s", sr.CommitHash, sr.Systems[platform].AttrPaths[0]), nil
}

func search(ctx context.Context, query string) (*SearchResults, error) {
	if query == "" {
		return nil, fn.Errorf("query should not be empty")
	}
	defer spinner.Client.UpdateMessage(fmt.Sprintf("searching for package %s", query))()

	endpoint, err := url.JoinPath(searchAPIEndpoint, "v1/search")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fn.Errorf("package %s not found", query)
		}

		return nil, fn.NewE(err)
	}
	url := endpoint + "?q=" + url.QueryEscape(query)

	return caller[SearchResults](ctx, url)
}

func (p *pkgHandler) syncPackage(pkghash string) error {
	if _, err := fn.Exec(fmt.Sprintf("nix shell nixpkgs/%s --command echo downloaded", pkghash), nil); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func createSet[T comparable](v []T) []T {
	m := make(map[T]struct{}, len(v))
	result := make([]T, 0, len(v))

	for i := range v {
		if _, ok := m[v[i]]; !ok {
			m[v[i]] = struct{}{}
			result = append(result, v[i])
		}
	}

	return result
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
