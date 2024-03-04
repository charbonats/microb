package dockerfile

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/charbonats/microbuild/v1/config"
	"github.com/charbonats/microbuild/v1/utils"
)

func runStage(c *config.Config) string {
	dockerfile := fromFinal(c)
	dockerfile += copy(c)
	dockerfile += entrypoint(c)
	dockerfile += labels(utils.Union(defaulLabels, c.Labels))

	return dockerfile
}

func fromFinal(c *config.Config) string {
	line := "\n"
	line += fmt.Sprintf("FROM python:%s-slim\n", c.PythonVersion)
	line += "RUN useradd --uid=65532 --user-group --home-dir=/home/nonroot --create-home nonroot\n"
	line += "USER 65532:65532\n"
	return line
}

func labels(labels map[string]string) string {
	line := "\n"
	for k, v := range labels {
		line += fmt.Sprintf("LABEL %s=\"%s\"\n", k, v)
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
