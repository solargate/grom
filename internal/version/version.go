package version

// Version is the product semantic version. Overridden at link time via:
//
//	-ldflags "-X github.com/solargate/grom/internal/version.Version=…"
var Version = "dev"
