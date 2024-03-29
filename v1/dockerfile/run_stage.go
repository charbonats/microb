package dockerfile

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/charbonats/microbuild/v1/config"
	"github.com/charbonats/microbuild/v1/utils"
	"mvdan.cc/sh/v3/shell"
)

func runStage(c *config.Config, placeholders map[string]string) string {
	dockerfile := fromFinalStage(c)
	if c.Flavor == "debian" {
		dockerfile += installSystemDepsWithApt(c)
	} else if c.Flavor == "alpine" {
		dockerfile += installSystemDepsWithApk(c)
	} else {
		log.Fatalf("unsupported flavor: %s", c.Flavor)
	}
	dockerfile += createNonRootUser(c)
	dockerfile += copyFiles(c)
	dockerfile += addFiles(c)
	dockerfile += addEntrypointAndCommand(c)
	dockerfile += addEnvironmentVariables(c.Env, placeholders)
	dockerfile += addLabels(utils.Union(defaulLabels, c.Labels), placeholders)
	dockerfile += addAuthorsLabels(c)
	return dockerfile
}

func fromFinalStage(c *config.Config) string {
	line := "\n"
	image := fmt.Sprintf("python:%s", c.PythonVersion)
	switch c.Flavor {
	case "alpine":
		image += "-alpine"
	case "debian":
		image += "-slim"
	}
	line += fmt.Sprintf("FROM %s\n", image)
	return line
}

func installSystemDepsWithApt(c *config.Config) string {
	line := "\n"
	if len(c.SystemDeps) > 0 {
		line += "RUN apt-get update && apt-get install -y --no-install-recommends "
		for _, dep := range c.SystemDeps {
			line += fmt.Sprintf(" %s ", dep)
		}
		line += " && rm -rf /var/lib/apt/lists/*\n"
	}
	return line
}

func installSystemDepsWithApk(c *config.Config) string {
	line := "\n"
	if len(c.SystemDeps) > 0 {
		line += "RUN apk add --no-cache "
		for _, dep := range c.SystemDeps {
			line += fmt.Sprintf(" %s ", dep)
		}
		line += "\n"
	}
	return line
}

func createNonRootUser(c *config.Config) string {
	line := "\n"
	if c.Flavor == "alpine" {
		line += "RUN addgroup 65532 && adduser -u 65532 -G 65532 -h /home/nonroot -D nonroot\n"
	} else {
		line += "RUN useradd --uid=65532 --user-group --home-dir=/home/nonroot --create-home nonroot\n"
	}
	line += "USER 65532:65532\n"
	return line
}

func addEnvironmentVariables(envs map[string]string, placeholders map[string]string) string {
	if len(envs) == 0 {
		return ""
	}
	lines := []string{"\n"}
	for k, v := range envs {
		v, err := shell.Expand(v, func(key string) string {
			return placeholders[key]
		})
		if err != nil {
			log.Fatal(err)
		}
		lines = append(lines, fmt.Sprintf("ENV %s=%s", k, v))
	}
	return strings.Join(lines, "\n")
}

func addLabels(labels map[string]string, placeholders map[string]string) string {
	line := "\n"
	for k, v := range labels {
		v, err := shell.Expand(v, func(key string) string {
			return placeholders[key]
		})
		if err != nil {
			log.Fatal(err)
		}
		line += fmt.Sprintf("LABEL %s=\"%s\"\n", k, v)
	}
	return line
}

func addAuthorsLabels(c *config.Config) string {
	line := "\n"
	if len(c.Authors) > 0 {
		authors := make([]string, len(c.Authors))
		for idx, author := range c.Authors {
			if author.Email != "" {
				authors[idx] = fmt.Sprintf("%s <%s>", author.Name, author.Email)
			} else {
				authors[idx] = author.Name
			}
		}
		line += fmt.Sprintf("LABEL org.opencontainers.image.authors=\"%s\"", strings.Join(authors, ", "))
	}
	return line
}

func copyFiles(c *config.Config) string {
	line := "\n"
	line += "COPY --from=builder /root/.local /home/nonroot/.local\n"
	line += "ENV PATH=$PATH:/home/nonroot/.local/bin\n"
	if len(c.CopyFiles) > 0 {
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

func addFiles(c *config.Config) string {
	line := "\n"
	if len(c.AddFiles) > 0 {
		line += "\n"
		for _, f := range c.AddFiles {
			if f.Checksum != "" {
				line += fmt.Sprintf("ADD --checksum=%s %s %s\n", f.Checksum, f.Source, f.Destination)
			}
			line += fmt.Sprintf("ADD %s %s\n", f.Source, f.Destination)
		}
	}
	return line
}

func addEntrypointAndCommand(c *config.Config) string {
	line := "\n"
	if len(c.Entrypoint) > 0 {
		entrypoint, err := json.Marshal(c.Entrypoint)
		if err != nil {
			log.Fatal(err)
		}
		line += fmt.Sprintf("ENTRYPOINT %s\n", entrypoint)
	}
	if len(c.Command) > 0 {
		cmd, err := json.Marshal(c.Command)
		if err != nil {
			log.Fatal(err)
		}
		line += fmt.Sprintf("CMD %s\n", cmd)
	}
	return line
}
