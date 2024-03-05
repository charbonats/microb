package dockerfile

import (
	"github.com/charbonats/microbuild/v1/config"
)

const pipCacheMount = " --mount=type=cache,target=/root/.cache"

// Apt needs exclusive access to its data, so the caches use the option sharing=locked,
// which will make sure multiple parallel builds using the same cache mount will wait for
// each other and not access the same cache files at the same time.
// See https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md#example-cache-apt-packages
const aptCacheMount = " --mount=type=cache,target=/var/cache/apt,sharing=locked --mount=type=cache,target=/var/lib/apt,sharing=locked"
const apkCacheMount = " --mount=type=cache,target=/var/cache/apk,sharing=locked"
const sshMount = " --mount=type=ssh,required=true"

var defaultEnvs = map[string]string{
	"PIP_DISABLE_PIP_VERSION_CHECK": "1",
	"PIP_NO_WARN_SCRIPT_LOCATION":   "0",
	"PIP_USER":                      "1",
	"PYTHONPYCACHEPREFIX":           "$HOME/.pycache",
}

var defaulLabels = map[string]string{
	"org.opencontainers.image.description": "autogenerated by microb",
	"moby.buildkit.frontend":               "microb",
	"microb.version":                       "v1",
}

type Options struct {
	Placeholders       map[string]string
	RequirementsUseSsh bool
}

// Microb2Dockerfile translates a microb config into a Dockerfile.
func Microb2Dockerfile(
	c *config.Config,
	options *Options,
) string {
	if options == nil {
		options = &Options{}
	}
	dockerfile := buildStage(c, options)
	dockerfile += runStage(c, options)
	return dockerfile
}
