package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/charbonats/microbuild/v1/utils"
	"github.com/hashicorp/go-version"
)

// NewConfigFromFile creates a new Config from a file path and a target.
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

// NewConfigFromBytes creates a new Config from a byte array and a target.
// Byte array is expected to be UTF-8 encoded TOML data from a pyproject.toml file.
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
	dependencies := pyproject.Project.Dependencies
	if len(appConfig.Extras) > 0 {
		if appConfig.Requirements != "" {
			return nil, fmt.Errorf("requirements field is not allowed when using extras")
		}
		for _, extra := range appConfig.Extras {
			extraDeps, ok := pyproject.Project.OptionalDependencies[extra]
			if !ok {
				return nil, fmt.Errorf("extra %s not found in pyproject.toml", extra)
			}
			dependencies = append(dependencies, extraDeps...)
		}
	}
	config := Config{
		Name:                 pyproject.Project.Name,
		Authors:              pyproject.Project.Authors,
		PythonVersion:        pythonVersion,
		Entrypoint:           appConfig.Entrypoint,
		Command:              appConfig.Command,
		Env:                  appConfig.Env,
		Labels:               appConfig.Labels,
		BuildDeps:            appConfig.BuildDeps,
		SystemDeps:           appConfig.SystemDeps,
		Dependencies:         utils.Unique(dependencies),
		Requirements:         appConfig.Requirements,
		Indices:              appConfig.Indices,
		CopyFiles:            appConfig.CopyFiles,
		CopyFilesBeforeBuild: appConfig.CopyFilesBeforeBuild,
		AddFiles:             appConfig.AddFiles,
		AddFilesBeforeBuild:  appConfig.AddFilesBeforeBuild,
	}
	return &config, nil
}

// Config is a struct that represents a build config.
// A config is obtained from merging information found
// at the project level and the target level.
type Config struct {
	Name                 string            // Name of the project
	Authors              []Author          // Authors of the project
	PythonVersion        string            // Python version to use
	Entrypoint           []string          // Default command to run. Arguments provided to the container will be appended to this command.
	Command              []string          // Command to run when no arguments are provided. Command is concatenated with the entrypoint.
	Env                  map[string]string // Additional environment variables to add to the final image
	Labels               map[string]string // Addiional labels to add to the final image
	BuildDeps            []string          // Build dependencies (not installed in final image)
	SystemDeps           []string          // System dependencies (not installed during build, only installed in final image)
	Indices              []Index           // Extra index urls to use
	Dependencies         []string          // Dependencies to install
	Requirements         string            // Path to requirements file
	CopyFiles            []Copy            // Files to copy to the final image
	CopyFilesBeforeBuild []Copy            // Files to copy to the build context before building
	AddFiles             []Add             // Files to add to the final image
	AddFilesBeforeBuild  []Add             // Files to add to the build context before building
}

// Copy is a struct that represents a file copy operation.
// From is optional and can be used to specify a source outside of the build context.
// When From is omitted, the source is assumed to be a file or directory in the build context.
type Copy struct {
	From        string `toml:"from"`
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

// Add is a struct that represents a file add operation.
// Checksum is optional and can be used to verify the integrity of the file.
type Add struct {
	Checksum    string `toml:"checksum"`
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

// Index is a struct that represents a package index.
// Trust is optional and can be used to skip certificate verification.
// It is not recommended to use trust unless you are sure the index is owned by you or a trusted party.
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

// Project is a struct that represents a project section in a pyproject.toml file.
type Project struct {
	Name                 string              `toml:"name"`
	Authors              []Author            `toml:"authors"`
	Dependencies         []string            `toml:"dependencies"`
	OptionalDependencies map[string][]string `toml:"optional-dependencies"`
	RequiresPython       string              `toml:"requires-python"`
}

// Author is a struct that represents an author found in a pyproject.toml file.
type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

// Tool is a struct that represents a tool section in a pyproject.toml file.
// It only contains the microb section and is not a complete representation of the file.
type Tool struct {
	Microb Microb `toml:"microb"`
}

// Microb is a struct that represents a microb section in a pyproject.toml file.
// For now, it only contains a map of targets.
type Microb struct {
	Target map[string]MicrobTarget `toml:"target"`
}

// MicrobTarget is a struct that represents a build target.
// All fields are optional and will be filled with default values if omitted.
type MicrobTarget struct {
	Entrypoint           []string          `toml:"entrypoint"`
	Command              []string          `toml:"command"`
	PythonVersion        string            `toml:"python_version"`
	Requirements         string            `toml:"requirements"`
	Indices              []Index           `toml:"indices"`
	Extras               []string          `toml:"extras"`
	Env                  map[string]string `toml:"environment"`
	Labels               map[string]string `toml:"labels"`
	BuildDeps            []string          `toml:"build_deps"`
	SystemDeps           []string          `toml:"system_deps"`
	CopyFiles            []Copy            `toml:"copy_files"`
	CopyFilesBeforeBuild []Copy            `toml:"copy_files_before_build"`
	AddFiles             []Add             `toml:"add_files"`
	AddFilesBeforeBuild  []Add             `toml:"add_files_before_build"`
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
