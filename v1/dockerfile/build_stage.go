package dockerfile

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/charbonats/microbuild/v1/config"
	"github.com/charbonats/microbuild/v1/utils"
)

func buildStage(c *config.Config, options *Options) string {
	dockerfile := fromBuilder(c)
	if c.Flavor == "debian" {
		dockerfile += installBuildDepsApt(c, options.RequirementsUseSsh)
	} else if c.Flavor == "alpine" {
		dockerfile += installBuildDepsApk(c, options.RequirementsUseSsh)
	} else {
		log.Fatalf("unsupported flavor: %s", c.Flavor)
	}
	dockerfile += env(utils.Union(defaultEnvs, c.Env), options.Placeholders)
	dockerfile += copyBeforeBuild(c)
	dockerfile += addBeforeBuild(c)
	if c.Requirements != "" {
		dockerfile += installPythonDepsFromRequirements(c, options.RequirementsUseSsh)
	} else {
		dockerfile += installPythonDeps(c)
	}
	dockerfile += installPythonProject(c)
	dockerfile += clearCachedDataFromInstall(c)
	return dockerfile
}

func fromBuilder(c *config.Config) string {
	tag := c.PythonVersion
	if c.Flavor == "alpine" {
		tag += "-alpine"
	}
	line := fmt.Sprintf("FROM docker.io/python:%s AS builder\n", tag)
	return line
}

func updateBuildDeps(c *config.Config, requirementsUseSsh bool) []string {
	needJq := false
	needGit := false
	needOpenssh := false
	if len(c.Indices) > 0 {
		for _, index := range c.Indices {
			if index.UsernameSecret != "" || index.PasswordSecret != "" {
				needJq = true
				break
			}
		}
	}
	deps := make([]string, len(c.BuildDeps))
	copy(deps, c.BuildDeps)
	if needJq {
		if len(deps) == 0 {
			deps = append(deps, "jq")
		} else {
			found := false
			for _, dep := range deps {
				if dep == "jq" {
					found = true
					break
				}
			}
			if !found {
				deps = append(deps, "jq")
			}
		}
	}
	for _, d := range c.Dependencies {
		if strings.Contains(d, "git+") {
			needGit = true
		}
		if strings.Contains(d, "git+ssh") {
			needOpenssh = true
		}
	}
	if needGit || requirementsUseSsh {
		if len(deps) == 0 {
			deps = append(deps, "git")
		} else {
			found := false
			for _, dep := range deps {
				if dep == "git" {
					found = true
					break
				}
			}
			if !found {
				deps = append(deps, "git")
			}
		}
	}
	if needOpenssh || requirementsUseSsh {
		if len(deps) == 0 {
			deps = append(deps, "openssh-client")
		} else {
			found := false
			for _, dep := range deps {
				if dep == "openssh-client" {
					found = true
					break
				}
			}
			if !found {
				deps = append(deps, "openssh-client")
			}
		}
	}
	return deps
}

func installBuildDepsApt(c *config.Config, requirementsUseSsh bool) string {
	deps := updateBuildDeps(c, requirementsUseSsh)
	if len(deps) == 0 {
		return ""
	}
	line := fmt.Sprintf("RUN %s ", aptCacheMount)
	line += "apt-get update && apt-get install -y --no-install-recommends "
	line += strings.Join(deps, " ")
	return line
}

func installBuildDepsApk(c *config.Config, requirementsUseSsh bool) string {
	deps := updateBuildDeps(c, requirementsUseSsh)
	if len(deps) == 0 {
		return ""
	}
	line := fmt.Sprintf("RUN %s ", apkCacheMount)
	line += "apk add "
	line += strings.Join(deps, " ")
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
		replaceUser := ""
		replacePassword := ""
		if index.UsernameSecret != "" {
			userSecretFile := fmt.Sprintf("/run/secrets/%s", index.UsernameSecret)
			replaceUser = fmt.Sprintf("$(echo -n $(cat %s) | jq -sRr @uri)", userSecretFile)
			index.Username = "REPLACE_USER"
		}
		if index.PasswordSecret != "" {
			passSecretFile := fmt.Sprintf("/run/secrets/%s", index.PasswordSecret)
			replacePassword = fmt.Sprintf("$(echo -n $(cat %s) | jq -sRr @uri)", passSecretFile)
			index.Password = "REPLACE_PASSWORD"
		}

		if len(strings.TrimSpace(index.Username)) != 0 && len(strings.TrimSpace(index.Password)) == 0 {
			indexUrl.User = url.User(index.Username)
		}

		if len(strings.TrimSpace(index.Username)) != 0 && len(strings.TrimSpace(index.Password)) != 0 {
			indexUrl.User = url.UserPassword(index.Username, index.Password)
		}
		indexUrlString := indexUrl.String()
		if replaceUser != "" {
			indexUrlString = strings.Replace(indexUrlString, "REPLACE_USER", replaceUser, 1)
		}
		if replacePassword != "" {
			indexUrlString = strings.Replace(indexUrlString, "REPLACE_PASSWORD", replacePassword, 1)
		}
		indices += fmt.Sprintf(" --extra-index-url \"%s\"", indexUrlString)

		if index.Trust {
			indices += fmt.Sprintf(" --trusted-host \"%s\"", indexUrl.Host)
		}
	}

	return indices
}

func installPythonDeps(c *config.Config) string {
	if len(c.Dependencies) == 0 {
		return ""
	}
	line := "\n"
	line += fmt.Sprintf("RUN %s", pipCacheMount)
	if len(c.Indices) > 0 {
		for _, index := range c.Indices {
			if index.PasswordSecret != "" {
				line += fmt.Sprintf(" --mount=type=secret,id=%s", index.PasswordSecret)
			}
			if index.UsernameSecret != "" {
				line += fmt.Sprintf(" --mount=type=secret,id=%s", index.UsernameSecret)
			}
		}
	}
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

func installPythonDepsFromRequirements(c *config.Config, useSsh bool) string {
	line := "\n"
	line += fmt.Sprintf("COPY %s /requirements.txt", c.Requirements)
	line += "\n"
	line += fmt.Sprintf("RUN %s", pipCacheMount)
	if len(c.Indices) > 0 {
		for _, index := range c.Indices {
			if index.PasswordSecret != "" {
				line += fmt.Sprintf(" --mount=type=secret,id=%s", index.PasswordSecret)
			}
			if index.UsernameSecret != "" {
				line += fmt.Sprintf(" --mount=type=secret,id=%s", index.UsernameSecret)
			}
		}
	}
	if useSsh {
		line += sshMount
		line += " GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no'"
	}
	line += fmt.Sprintf(" python -m pip install --user %s -r /requirements.txt", indices(c))
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
