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
	dockerfile := fromFinal(c)
	dockerfile += installSystemDeps(c)
	dockerfile += nonRootUser(c)
	dockerfile += copy(c)
	dockerfile += entrypoint(c)
	dockerfile += env(c.Env, placeholders)
	dockerfile += labels(utils.Union(defaulLabels, c.Labels), placeholders)
	dockerfile += authors(c)
	return dockerfile
}

func fromFinal(c *config.Config) string {
	line := "\n"
	line += fmt.Sprintf("FROM python:%s-slim\n", c.PythonVersion)
	return line
}

func installSystemDeps(c *config.Config) string {
	line := "\n"
	if len(c.SystemDeps) > 0 {
		line += "RUN apt-get update && apt-get install -y --no-install-recommends "
		for _, dep := range c.SystemDeps {
			line += fmt.Sprintf(" %s ", dep)
		}
		line += " && rm -rf /var/lib/apt/lists/*"
	}
	return line
}

func nonRootUser(c *config.Config) string {
	line := "\n"
	line += "RUN useradd --uid=65532 --user-group --home-dir=/home/nonroot --create-home nonroot\n"
	line += "USER 65532:65532\n"
	return line
}

func env(envs map[string]string, placeholders map[string]string) string {
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

func labels(labels map[string]string, placeholders map[string]string) string {
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

func authors(c *config.Config) string {
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

func copy(c *config.Config) string {
	line := "\n"
	if len(c.Dependencies) > 0 {
		line += "COPY --from=builder /root/.local /home/nonroot/.local\n"
		line += "ENV PATH=$PATH:/home/nonroot/.local/bin\n"
	}
	if len(c.CopyFiles) > 0 {
		line += "\n"
		for _, f := range c.CopyFiles {
			line += fmt.Sprintf("COPY %s %s\n", f.Source, f.Destination)
		}
	}
	return line
}

func entrypoint(c *config.Config) string {
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
