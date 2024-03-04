package dockerfile

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/charbonats/microbuild/v1/config"
	"github.com/charbonats/microbuild/v1/utils"
)

func buildStage(c *config.Config, placeholders map[string]string) string {
	dockerfile := fromBuilder(c)
	dockerfile += installBuildDeps(c)
	dockerfile += env(utils.Union(defaultEnvs, c.Env), placeholders)
	dockerfile += copyBeforeBuild(c)
	dockerfile += addBeforeBuild(c)
	dockerfile += installPythonDeps(c)
	dockerfile += installPythonProject(c)
	dockerfile += clearCachedDataFromInstall(c)
	return dockerfile
}

func fromBuilder(c *config.Config) string {
	line := fmt.Sprintf("FROM docker.io/python:%s AS builder\n", c.PythonVersion)
	return line
}

func installBuildDeps(c *config.Config) string {
	if len(c.BuildDeps) == 0 {
		return ""
	}
	line := fmt.Sprintf("RUN %s ", aptCacheMount)
	line += "apt-get update && apt-get install -y --no-install-recommends "
	line += strings.Join(c.BuildDeps, " ")
	return line
}

func copyBeforeBuild(c *config.Config) string {
	line := ""
	if len(c.CopyFilesBeforeBuild) > 0 {
		line += "\n"
		for _, f := range c.CopyFiles {
			if f.From != "" {
				line += fmt.Sprintf("COPY --from=%s %s %s\n", f.From, f.Source, f.Destination)
			} else {
				line += fmt.Sprintf("COPY %s %s\n", f.Source, f.Destination)
			}
		}
	}
	return line
}

func addBeforeBuild(c *config.Config) string {
	line := ""
	if len(c.AddFilesBeforeBuild) > 0 {
		line += "\n"
		for _, f := range c.AddFilesBeforeBuild {
			if f.Checksum != "" {
				line += fmt.Sprintf("ADD --checksum=%s %s %s\n", f.Checksum, f.Source, f.Destination)
			}
			line += fmt.Sprintf("ADD %s %s\n", f.Source, f.Destination)
		}
	}
	return line
}

func indices(c *config.Config) string {
	indices := "--retries 2"

	for _, index := range c.Indices {
		indexUrl, err := url.Parse(index.Url)
		if err != nil {
			log.Fatal(err)
		}

		if len(strings.TrimSpace(index.Username)) != 0 && len(strings.TrimSpace(index.Password)) == 0 {
			indexUrl.User = url.User(index.Username)
		}

		if len(strings.TrimSpace(index.Username)) != 0 && len(strings.TrimSpace(index.Password)) != 0 {
			indexUrl.User = url.UserPassword(index.Username, index.Password)
		}

		indices += fmt.Sprintf(" --extra-index-url %s", indexUrl.String())

		if index.Trust {
			indices += fmt.Sprintf(" --trusted-host %s", indexUrl.Host)
		}
	}

	return indices
}

func installPythonDeps(c *config.Config) string {
	line := "\n"
	line += fmt.Sprintf("RUN %s", pipCacheMount)
	useSsh := false
	for _, d := range c.Dependencies {
		if strings.Contains(d, "git+ssh") {
			useSsh = true
			break
		}
	}
	if useSsh {
		line += sshMount
		line += " GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no'"
	}
	line += fmt.Sprintf(" python -m pip install --user %s ", indices(c))
	line += strings.Join(c.Dependencies, " ")
	return line
}

func installPythonProject(c *config.Config) string {
	line := "\n"
	line += "COPY . /projectdir\n"
	line += fmt.Sprintf("RUN %s python -m pip install --no-deps /projectdir", pipCacheMount)
	return line
}

func clearCachedDataFromInstall(c *config.Config) string {
	line := "\n"
	if len(c.Dependencies) > 0 {
		line += "RUN find /root/.local/lib/python*/ -name 'tests' -exec rm -r '{}' + && "
		line += "find /root/.local/lib/python*/site-packages/ -name '*.so' -exec sh -c 'file \"{}\" | grep -q \"not stripped\" && strip -s \"{}\"' \\; && "
		line += "find /root/.local/lib/python*/ -type f -name '*.pyc' -delete && "
		line += "find /root/.local/lib/python*/ -type d -name '__pycache__' -delete\n"
	}

	return line
}
