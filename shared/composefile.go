package shared

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ComposefileName  = "brave-compose.yaml"
	ComposefileAlias = "brave-compose.yml"
)

// ComposeService defines a service
type ComposeService struct {
	Service        `yaml:",inline"`
	BravefileBuild *Bravefile
	Bravefile      string   `yaml:"bravefile,omitempty"`
	Build          bool     `yaml:"build,omitempty"`
	Base           bool     `yaml:"base,omitempty"`
	Context        string   `yaml:"context,omitempty"`
	Depends        []string `yaml:"depends_on,omitempty"`
}

// A ComposeFile maps service names to services
type ComposeFile struct {
	Path     string
	Services map[string]*ComposeService `yaml:"services"`
}

// NewComposeFile returns a pointer to a newly created empty ComposeFile struct
func NewComposeFile() *ComposeFile {
	return &ComposeFile{}
}

// Load reads a compose file from disk and loads its settings into the composeFile struct
func (composeFile *ComposeFile) Load(file string) error {
	buf, err := ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf.Bytes(), &composeFile)
	if err != nil {
		return err
	}

	// Record composefile path (later used for deploy context)
	composeFile.Path = file

	// Check for empty compose file
	if len(composeFile.Services) == 0 {
		return fmt.Errorf("no services found in composefile %q", composeFile.Path)
	}

	// Move to parent directory of compose file so relative Bravefile paths work
	workingDir, err := filepath.Abs(filepath.Dir(composeFile.Path))
	if err != nil {
		return err
	}
	startDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(workingDir)
	defer os.Chdir(startDir)

	// Upade each service with servicename and load bravefile if provided
	for serviceName := range composeFile.Services {
		service := composeFile.Services[serviceName]

		// Override Service.Name with the key provided in brave-compose file
		service.Name = serviceName

		if (service.Build || service.Base) && service.Bravefile == "" {
			return fmt.Errorf("cannot build image for %q without a Bravefile path", service.Name)
		}

		// Load Bravefile is provided - merge service settings and save build settings
		if service.Bravefile != "" {
			service.BravefileBuild = NewBravefile()
			err = service.BravefileBuild.Load(service.Bravefile)
			if err != nil {
				return fmt.Errorf("failed to load bravefile %q", service.Bravefile)
			}
			service.Service.Merge(&service.BravefileBuild.PlatformService)
		}
	}

	return nil
}

// TopologicalOrdering returns a string array of service names that are ordered
// so that each service comes after the services it depends on.
// If a valid ordering cannot be found due to cycles in the graph an error will be returned.
func (composeFile *ComposeFile) TopologicalOrdering() (topologicalOrdering []string, err error) {

	// digraph maps services to dependents
	// outdegrees maps services to unfulfilled dependency count
	digraph := make(map[string][]string, len(composeFile.Services))
	outdegrees := make(map[string]int, len(composeFile.Services))

	for service := range composeFile.Services {
		digraph[service] = nil
		outdegrees[service] = 0
	}

	for service := range composeFile.Services {
		for _, dependency := range composeFile.Services[service].Depends {
			_, exists := outdegrees[dependency]
			if !exists {
				return topologicalOrdering, fmt.Errorf("service %q depends on service %q which does not exist", service, dependency)
			}
			digraph[dependency] = append(digraph[dependency], service)
			outdegrees[service] += 1
		}
	}

	for progress := true; progress; {
		progress = false
		for service := range outdegrees {
			// take service with 0 remaining unfulfilled dependencies
			if outdegrees[service] == 0 {
				topologicalOrdering = append(topologicalOrdering, service)
				delete(outdegrees, service)
				// update dependent count for any services depending on this service
				for _, dependent := range digraph[service] {
					outdegrees[dependent] -= 1
				}
				progress = true
			}
		}
	}

	if len(topologicalOrdering) < len(composeFile.Services) {
		return nil, errors.New("no valid topological sorting of dependency graph found - check graph for cycles")
	}

	return topologicalOrdering, nil
}
