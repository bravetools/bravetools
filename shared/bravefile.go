package shared

import (
	"gopkg.in/yaml.v2"
)

// ImageDescription defines base image type and source
type ImageDescription struct {
	Image    string `yaml:"image"`
	Location string `yaml:"location"`
}

// Packages defines system packages to install in container
type Packages struct {
	Manager string   `yaml:"manager,omitempty"`
	System  []string `yaml:"system,omitempty"`
}

// RunCommand defines custom commands to run inside continer
type RunCommand struct {
	Command string   `yaml:"command,omitempty"`
	Content string   `yaml:"content,omitempty"`
	Args    []string `yaml:"args,omitempty"`
}

//CopyCommand defines source and target for files to be copied into container
type CopyCommand struct {
	Source string `yaml:"source,omitempty"`
	Target string `yaml:"target,omitempty"`
	Action string `yaml:"action,omitempty"`
}

// Service defines command to install app
type Service struct {
	Name       string     `yaml:"name,omitempty"`
	Image      string     `yaml:"image,omitempty"`
	Version    string     `yaml:"version,omitempty"`
	Docker     string     `yaml:"docker,omitempty"`
	IP         string     `yaml:"ip"`
	Ports      []string   `yaml:"ports"`
	Resources  Resources  `yaml:"resources"`
	Postdeploy Postdeploy `yaml:"postdeploy,omitempty"`
}

// Postdeploy defines operations to perform after service deployment finish
type Postdeploy struct {
	Run  []RunCommand  `yaml:"run,omitempty"`
	Copy []CopyCommand `yaml:"copy,omitempty"`
}

// Resources defines resources allocated to service
type Resources struct {
	RAM string `yaml:"ram"`
	CPU string `yaml:"cpu"`
	GPU string `yaml:"gpu"`
}

// Bravefile describes unit configuration
type Bravefile struct {
	Base            ImageDescription `yaml:"base"`
	SystemPackages  Packages         `yaml:"packages,omitempty"`
	Run             []RunCommand     `yaml:"run,omitempty"`
	Copy            []CopyCommand    `yaml:"copy,omitempty"`
	PlatformService Service          `yaml:"service,omitempty"`
}

// NewBravefile ..
func NewBravefile() *Bravefile {
	return &Bravefile{}
}

// Load loads Bravefile
func (bravefile *Bravefile) Load(file string) error {

	buf, err := ReadFile(file)
	if err != nil {
		return err
	}

	bravefile.PlatformService.Resources.CPU = DefaultUnitCpuLimit
	bravefile.PlatformService.Resources.RAM = DefaultUnitRamLimit

	err = yaml.Unmarshal(buf.Bytes(), &bravefile)
	if err != nil {
		return err
	}

	return nil
}
