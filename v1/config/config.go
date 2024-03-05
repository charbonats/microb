package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charbonats/microbuild/v1/utils"
)

// Options is a struct that represents options for the build process.
// Options are deduced from the build context, not from the pyproject.toml file.
type Options struct {
	Filename          string
	Target            string
	BuildArgs         map[string]string
	ReadRequirements  func(name string) ([]string, error)
	ReadPythonVersion func() string
}

// NewConfigFromFile creates a new Config from a file path and a target.
func NewConfigFromFile(path string, options *Options) (*Config, error) {
	content, err := utils.ReadFileAsBytes(path)
	if err != nil {
		return nil, fmt.Errorf("NewConfigFromFile: %w", err)
	}
	return NewConfigFromBytes(content, options)
}

// NewConfigFromBytes creates a new Config from a byte array and a target.
// Byte array is expected to be UTF-8 encoded TOML data from a pyproject.toml file.
func NewConfigFromBytes(data []byte, options *Options) (*Config, error) {
	var pyproject PyProject
	// Start by decoding the pyproject.toml file
	_, err := toml.Decode(string(data), &pyproject)
	if err != nil {
		return nil, fmt.Errorf("NewConfigFromBytes: failed to decode pyproject.toml content: %w", err)
	}
	// Get the constraints on Python versions by the project
	requiresPython := pyproject.Project.RequiresPython
	// If we're using poetry, we need to check the python version constraints from there
	if pyproject.Tool.Poetry.Name != "" {
		requiresPython = pyproject.Tool.Poetry.PythonRequires()
	}
	target := options.Target
	// If no target is specified
	if target == "" {
		// Look for the first target in the microb config
		defaultTarget, ok := defaultTarget(&pyproject.Tool.Microb)
		// If there is still no target found, use default values
		if !ok {
			pythonVersion, err := GetPythonVersion(requiresPython, options.ReadPythonVersion())
			if err != nil {
				return nil, err
			}
			dependenciesUseSsh := isUsingSsh(pyproject.Project.Dependencies)
			dependenciesUseGit := isUsingGit(pyproject.Project.Dependencies)
			return &Config{
				Flavor:             DefaultFlavor(),
				Name:               pyproject.Project.Name,
				Authors:            pyproject.Project.Authors,
				PythonVersion:      pythonVersion,
				Dependencies:       pyproject.Project.Dependencies,
				DependenciesUseSsh: dependenciesUseSsh,
				DependenciesUseGit: dependenciesUseGit,
			}, nil
			// Else use the first target found
		} else {
			target = defaultTarget
		}
	}
	// Get the target config
	targetConfig, ok := pyproject.Tool.Microb.Target[target]
	if !ok {
		return nil, fmt.Errorf("NewConfigFromBytes: target %s not found in pyproject.toml", target)
	}
	// Validate the build flavor
	targetConfig.Flavor, ok = Flavor(targetConfig.Flavor)
	if !ok {
		return nil, fmt.Errorf("NewConfigFromBytes: target %s uses unknown flavor %s", target, targetConfig.Flavor)
	}
	// If no python version is specified, use the default
	if targetConfig.PythonVersion == "" {
		targetConfig.PythonVersion = options.ReadPythonVersion()
	}
	// Validate the python version
	pythonVersion, err := GetPythonVersion(requiresPython, targetConfig.PythonVersion)
	if err != nil {
		return nil, fmt.Errorf("NewConfigFromBytes: failed to get python verson for target %s: %w", target, err)
	}
	if targetConfig.Requirements != "" && len(targetConfig.Extras) > 0 {
		return nil, fmt.Errorf("NewConfigFromBytes: failed to validate configuration for taget %s: using requirements is not allowed together with extras", target)
	}
	// Merge the dependencies with extras if any
	dependencies, err := getPythonDeps(&pyproject, targetConfig.Extras)
	if err != nil {
		return nil, fmt.Errorf("NewConfigFromBytes: failed to get dependencies for target %s: %w", target, err)
	}
	dependenciesUseSsh := false
	dependenciesUseGit := false
	if targetConfig.Requirements != "" {
		reqs, err := options.ReadRequirements(targetConfig.Requirements)
		if err != nil {
			return nil, fmt.Errorf("NewConfigFromBytes: failed to get requirements for target %s: %w", target, err)
		}
		dependenciesUseSsh = isUsingSsh(reqs)
		dependenciesUseGit = isUsingGit(reqs)
	} else {
		dependenciesUseSsh = isUsingSsh(dependencies)
		dependenciesUseGit = isUsingGit(dependencies)
	}
	buildDeps := getBuildDeps(targetConfig.Indices, targetConfig.BuildDeps, dependenciesUseSsh, dependenciesUseGit)
	config := Config{
		Flavor:               targetConfig.Flavor,
		Name:                 pyproject.Project.Name,
		Authors:              pyproject.Project.Authors,
		PythonVersion:        pythonVersion,
		Entrypoint:           targetConfig.Entrypoint,
		Command:              targetConfig.Command,
		Env:                  targetConfig.Env,
		Labels:               targetConfig.Labels,
		BuildDeps:            buildDeps,
		SystemDeps:           targetConfig.SystemDeps,
		Dependencies:         dependencies,
		Requirements:         targetConfig.Requirements,
		DependenciesUseSsh:   dependenciesUseSsh,
		DependenciesUseGit:   dependenciesUseGit,
		Indices:              targetConfig.Indices,
		CopyFiles:            targetConfig.CopyFiles,
		CopyFilesBeforeBuild: targetConfig.CopyFilesBeforeBuild,
		AddFiles:             targetConfig.AddFiles,
		AddFilesBeforeBuild:  targetConfig.AddFilesBeforeBuild,
	}
	return &config, nil
}

// Config is a struct that represents a build config.
// A config is obtained from merging information found
// at the project level and the target level.
type Config struct {
	Flavor               string            // Flavor of the build ("debian" or "alpine")
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
	DependenciesUseSsh   bool              // Whether ssh is required to install dependencies or not
	DependenciesUseGit   bool              // Whether git is required to install dependencies or not
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
	Url            string `toml:"url"`
	Username       string `toml:"username"`
	UsernameSecret string `toml:"username_secret"`
	Password       string `toml:"password"`
	PasswordSecret string `toml:"password_secret"`
	Trust          bool   `toml:"trust"`
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
	Poetry Poetry `toml:"poetry"`
}

// Microb is a struct that represents a microb section in a pyproject.toml file.
// For now, it only contains a map of targets.
type Microb struct {
	Target map[string]MicrobTarget `toml:"target"`
}

// MicrobTarget is a struct that represents a build target.
// All fields are optional and will be filled with default values if omitted.
type MicrobTarget struct {
	Flavor               string            `toml:"flavor"`
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

func getBuildDeps(
	indices []Index,
	buildDeps []string,
	dependenciesUseSsh bool,
	dependenciesUseGit bool,
) []string {
	deps := make([]string, len(buildDeps))
	copy(deps, buildDeps)
	if dependenciesUseSsh {
		deps = append(deps, "openssh-client")
	}
	if dependenciesUseGit {
		deps = append(deps, "git")
	}
	needJq := false
	if len(indices) > 0 {
		for _, index := range indices {
			if index.UsernameSecret != "" || index.PasswordSecret != "" {
				needJq = true
				break
			}
		}
	}
	if needJq {
		deps = append(deps, "jq")
	}
	return deps
}

func getPythonDeps(pyproject *PyProject, extras []string) ([]string, error) {
	dependencies := make([]string, len(pyproject.Project.Dependencies))
	copy(dependencies, pyproject.Project.Dependencies)
	if len(extras) > 0 {
		for _, extra := range extras {
			extraDeps, ok := pyproject.Project.OptionalDependencies[extra]
			if !ok {
				return nil, fmt.Errorf("extra %s not found in pyproject.toml", extra)
			}
			dependencies = append(dependencies, extraDeps...)
		}
	}
	return utils.Unique(dependencies), nil
}

func isUsingSsh(requirements []string) bool {
	for _, line := range requirements {
		if strings.Contains(line, "git+ssh://") {
			return true
		}
	}
	return false
}

func isUsingGit(requirements []string) bool {
	for _, line := range requirements {
		if strings.Contains(line, "git+") {
			return true
		}
	}
	return false
}

// DefaultTarget returns the first target found in the microb section.
func defaultTarget(m *Microb) (string, bool) {
	for name := range m.Target {
		return name, true
	}
	return "", false
}
