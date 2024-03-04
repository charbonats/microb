package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/go-version"
)

func NewConfigFromBytes(data []byte, target string) (*Config, error) {
	var pyproject PyProject
	_, err := toml.Decode(string(data), &pyproject)
	if err != nil {
		return nil, err
	}
	requiresPython := pyproject.Project.RequiresPython
	// If no target is specified
	if target == "" {
		// Look for the first target in the microb config
		for name := range pyproject.Tool.Microb.Target {
			target = name
			break
		}
		// If there is still no target found, use default values
		if target == "" {
			pythonVersion, err := findVersion(requiresPython, "")
			if err != nil {
				return nil, err
			}
			return &Config{
				Name:          pyproject.Project.Name,
				Authors:       pyproject.Project.Authors,
				PythonVersion: pythonVersion,
				Dependencies:  pyproject.Project.Dependencies,
			}, nil
		}
	}
	appConfig, ok := pyproject.Tool.Microb.Target[target]
	if !ok {
		return nil, fmt.Errorf("target %s not found in pyproject.toml", target)
	}
	pythonVersion, err := findVersion(requiresPython, appConfig.PythonVersion)
	if err != nil {
		return nil, err
	}
	config := Config{
		Name:          pyproject.Project.Name,
		Authors:       pyproject.Project.Authors,
		PythonVersion: pythonVersion,
		Entrypoint:    appConfig.Entrypoint,
		Command:       appConfig.Command,
		Env:           appConfig.Env,
		Labels:        appConfig.Labels,
		BuildDeps:     appConfig.BuildDeps,
		SystemDeps:    appConfig.SystemDeps,
		Dependencies:  pyproject.Project.Dependencies,
		Indices:       appConfig.Indices,
		CopyFiles:     appConfig.CopyFiles,
	}
	return &config, nil
}

func NewConfigFromFile(path string, target string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return NewConfigFromBytes(content, target)
}

// Config is a struct that represents a build config
type Config struct {
	Name          string            // Name of the project
	Authors       []Author          // Authors of the project
	PythonVersion string            // Python version to use
	Entrypoint    []string          // Default command to run. Arguments provided to the container will be appended to this command.
	Command       []string          // Command to run when no arguments are provided. Command is concatenated with the entrypoint.
	Env           map[string]string // Additional environment variables to add to the final image
	Labels        map[string]string // Addiional labels to add to the final image
	BuildDeps     []string          // Build dependencies (not installed in final image)
	SystemDeps    []string          // System dependencies (not installed during build, only installed in final image)
	Indices       []Index           // Extra index urls to use
	Dependencies  []string          // Dependencies to install
	CopyFiles     []FileToCopy      // Files to copy to the final image
}

type FileToCopy struct {
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

type Index struct {
	Url      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Trust    bool   `toml:"trust"`
}

// PyProject is a struct that represents a pyproject.toml file (partially)
type PyProject struct {
	Project Project `toml:"project"`
	Tool    Tool    `toml:"tool"`
}

type Project struct {
	Name           string   `toml:"name"`
	Authors        []Author `toml:"authors"`
	Dependencies   []string `toml:"dependencies"`
	RequiresPython string   `toml:"requires-python"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type Tool struct {
	Microb Microb `toml:"microb"`
}

type Microb struct {
	Target map[string]MicrobTarget `toml:"target"`
}

type MicrobTarget struct {
	Entrypoint    []string          `toml:"entrypoint"`
	Command       []string          `toml:"command"`
	PythonVersion string            `toml:"python_version"`
	Indices       []Index           `toml:"indices"`
	Env           map[string]string `toml:"environment"`
	Labels        map[string]string `toml:"labels"`
	BuildDeps     []string          `toml:"build_deps"`
	SystemDeps    []string          `toml:"system_deps"`
	CopyFiles     []FileToCopy      `toml:"copy_files"`
}

func (c Config) Export() (string, error) {
	var out bytes.Buffer
	json.NewEncoder(&out).Encode(c)
	return out.String(), nil
}

func findVersion(requires string, target string) (string, error) {
	constraints, err := version.NewConstraint(requires)
	if err != nil {
		return "", err
	}
	if target != "" {
		v, err := version.NewVersion(target)
		if err != nil {
			return "", err
		}
		if constraints.Check(v) {
			return target, nil
		} else {
			return "", fmt.Errorf("version %s does not satisfy the requirement %s", target, requires)
		}
	}
	for _, target := range []string{"3.11", "3.10", "3.9", "3.8", "3.7", "3.6"} {
		v, err := version.NewVersion(target)
		if err != nil {
			log.Fatal(err)
		}
		if constraints.Check(v) {
			return target, nil
		}
	}
	return "", fmt.Errorf("no version satisfies the requirement %s", requires)
}
