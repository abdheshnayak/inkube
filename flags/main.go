package flags

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
)

var (
	Version = "development"
	DevMode = "false"

	IsVerbose = false
	IsQuiet   = false

	CacheHome = xdg.CacheHome
	CacheDir  = fmt.Sprintf("%s/inkube", CacheHome)
)

func IsDev() bool {
	if DevMode == "false" {
		return false
	}
	return true
}

func GetCacheDir() string {
	os.MkdirAll(CacheDir, 0755)
	return CacheDir
}
