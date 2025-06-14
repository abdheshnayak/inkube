package nixpkghandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	fn "github.com/abdheshnayak/inkube/pkg/functions"
)

type PackageInfo struct {
	ID           int      `json:"id"`
	CommitHash   string   `json:"commit_hash"`
	System       string   `json:"system"`
	LastUpdated  int      `json:"last_updated"`
	StoreHash    string   `json:"store_hash"`
	StoreName    string   `json:"store_name"`
	StoreVersion string   `json:"store_version"`
	MetaName     string   `json:"meta_name"`
	MetaVersion  []string `json:"meta_version"`
	AttrPaths    []string `json:"attr_paths"`
	Version      string   `json:"version"`
	Summary      string   `json:"summary"`
}

type PackageVersion struct {
	PackageInfo

	Name    string                 `json:"name"`
	Systems map[string]PackageInfo `json:"systems,omitempty"`
}

type Package struct {
	Name        string           `json:"name"`
	NumVersions int              `json:"num_versions"`
	Versions    []PackageVersion `json:"versions,omitempty"`
}

type SearchResults struct {
	NumResults int       `json:"num_results"`
	Packages   []Package `json:"packages,omitempty"`
}

const searchAPIEndpoint = "https://search.devbox.sh"

var ErrNotFound = fn.Error("not found")

func caller[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fn.Errorf("GET %s: %w", url, err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fn.Errorf("GET %s: %w", url, err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fn.Errorf("GET %s: read respoonse body: %w", url, err)
	}

	if response.StatusCode == 404 {
		return nil, ErrNotFound
	}

	if response.StatusCode >= 400 {
		return nil, fn.Errorf("GET %s: unexpected status code %s: %s",
			url,
			response.Status,
			data,
		)
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fn.Errorf("GET %s: unmarshal response JSON: %w", url, err)
	}
	return &result, nil
}
