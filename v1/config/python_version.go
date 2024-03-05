package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	ALLOWED_PYTHON_VERSIONS = []string{"3.12", "3.11", "3.10", "3.9", "3.8", "3.7", "3.6"}
)

func GetPythonVersion(requires string, candidate string) (string, error) {
	// When we read version from file, there might be a leading line break
	requires = strings.TrimSpace(strings.Split(requires, "\n")[0])
	candidate = strings.TrimSpace(strings.Split(candidate, "\n")[0])
	constraints, err := version.NewConstraint(requires)
	if err != nil {
		return "", err
	}
	if candidate != "" {
		v, err := version.NewVersion(candidate)
		if err != nil {
			return "", fmt.Errorf("GetPythonVersion: version %s is not valid: %w", candidate, err)
		}
		if constraints.Check(v) {
			return candidate, nil
		} else {
			return "", fmt.Errorf("GetPythonVersion: version %s does not satisfy the constraint %s", candidate, requires)
		}
	}
	for _, target := range ALLOWED_PYTHON_VERSIONS {
		v, err := version.NewVersion(target)
		if err != nil {
			log.Fatal(err)
		}
		if constraints.Check(v) {
			return target, nil
		}
	}
	return "", fmt.Errorf("GetPythonVersion: no version satisfies the constraint %s", requires)
}
