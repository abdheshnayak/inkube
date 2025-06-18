package flags

var (
	Version = "development"
	DevMode = "false"

	IsVerbose = false
	IsQuiet   = false
)

func IsDev() bool {
	if DevMode == "false" {
		return false
	}
	return true
}
