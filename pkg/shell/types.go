package shell

import (
	"io"
	"slices"
	"strings"
)

// Naming Convention:
// - suffix Opts for structs corresponding to a Inkube api function
// - omit suffix Opts for other structs that are composed into an Opts struct

type Opts struct {
	Dir                      string
	Env                      map[string]string
	Environment              string
	IgnoreWarnings           bool
	CustomProcessComposeFile string
	Stderr                   io.Writer
}

type ProcessComposeOpts struct {
	ExtraFlags         []string
	Background         bool
	ProcessComposePort int
}

type GenerateOpts struct {
	ForType  string
	Force    bool
	RootUser bool
}

type EnvFlags struct {
	EnvMap  map[string]string
	EnvFile string
}

type PullboxOpts struct {
	Overwrite   bool
	URL         string
	Credentials Credentials
}

type Credentials struct {
	IDToken string
	// TODO We can just parse these out, but don't want to add a dependency right now
	Email string
	Sub   string
}

type AddOpts struct {
	AllowInsecure    []string
	Platforms        []string
	ExcludePlatforms []string
	DisablePlugin    bool
	Patch            string
	Outputs          []string
}

type UpdateOpts struct {
	Pkgs                  []string
	NoInstall             bool
	IgnoreMissingPackages bool
}

type EnvExportsOpts struct {
	EnvOptions     EnvOptions
	NoRefreshAlias bool
	RunHooks       bool
}

// EnvOptions configure the Inkube Environment in the `computeEnv` function.
// - These options are commonly set by flags in some Inkube commands
// like `shellenv`, `shell` and `run`.
// - The struct is designed for the "common case" to be zero-initialized as `EnvOptions{}`.
type EnvOptions struct {
	Hooks             LifecycleHooks
	OmitNixEnv        bool
	PreservePathStack bool
	Pure              bool
	SkipRecompute     bool
}

type LifecycleHooks struct {
	// OnStaleState is called when the Inkube state is out of date
	OnStaleState func()
}

func MapToPairs(m map[string]string) []string {
	pairs := make([]string, len(m))
	i := 0
	for k, v := range m {
		pairs[i] = k + "=" + v
		i++
	}
	slices.Sort(pairs) // for reproducibility
	return pairs
}

// PairsToMap creates a map from a slice of "key=value" environment variable
// pairs. Note that maps are not ordered, which can affect the final variable
// values when pairs contains duplicate keys.
func PairsToMap(pairs []string) map[string]string {
	vars := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			continue
		}
		vars[k] = v
	}
	return vars
}
