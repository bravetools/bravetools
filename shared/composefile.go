package shared

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

// ComposeService defines a service
type ComposeService struct {
	Service   `yaml:",inline"`
	Bravefile string   `yaml:"bravefile,omitempty"`
	Build     bool     `yaml:"build,omitempty"`
	Depends   []string `yaml:"depends_on,omitempty"`
}

// A ComposeFile maps service names to services
type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
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
