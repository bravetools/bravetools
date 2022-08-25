package shared

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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
	Command string            `yaml:"command,omitempty"`
	Content string            `yaml:"content,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
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
	return &Bravefile{
		PlatformService: Service{
			Resources: Resources{
				CPU: DefaultUnitCpuLimit,
				RAM: DefaultUnitRamLimit,
			},
		},
	}
}

// Load loads Bravefile
func (bravefile *Bravefile) Load(file string) error {

	buf, err := ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf.Bytes(), &bravefile)
	if err != nil {
		return err
	}

	return nil
}

// Validate validates Bravefile
func (bravefile *Bravefile) Validate() error {
	if bravefile.Base.Image == "" {
		return errors.New("invalid Bravefile: empty Base Image name")
	}

	if bravefile.PlatformService.Name == "" {
		return errors.New("invalid Bravefile: empty Service Name")
	}

	if bravefile.PlatformService.Image == "" {
		return errors.New("invalid Bravefile: empty Service Image name")
	}

	return nil
}

// Merges two Service structs, prioritizing the values present in first struct
func (s *Service) Merge(service *Service) {
	if s.Name == "" {
		s.Name = service.Name
	}
	if s.Image == "" {
		s.Image = service.Image
	}
	if s.Version == "" {
		s.Version = service.Version
	}
	if s.Docker == "" {
		s.Docker = service.Docker
	}
	if s.IP == "" {
		s.IP = service.IP
	}
	if len(s.Ports) == 0 {
		s.Ports = append(s.Ports, service.Ports...)
	}
	if s.Resources.CPU == "" {
		s.Resources.CPU = service.Resources.CPU
	}
	if s.Resources.GPU == "" {
		s.Resources.GPU = service.Resources.GPU
	}
	if s.Resources.RAM == "" {
		s.Resources.RAM = service.Resources.RAM
	}
	if len(s.Postdeploy.Copy) == 0 {
		s.Postdeploy.Copy = append(s.Postdeploy.Copy, service.Postdeploy.Copy...)
	}
	if len(s.Postdeploy.Run) == 0 {
		s.Postdeploy.Run = append(s.Postdeploy.Run, service.Postdeploy.Run...)
	}
}

// GetBravefileFromGitHub reads bravefile from a github URL
func GetBravefileFromGitHub(name string) (*Bravefile, error) {
	var bravefile Bravefile
	var baseConfig string

	path := strings.SplitN(name, "/", -1)
	if len(path) <= 3 {
		return nil, fmt.Errorf("failed to retrieve image %q from github", name)
	}
	user := path[1]
	repository := path[2]
	project := strings.Join(path[3:], "/")

	url := "https://raw.githubusercontent.com/" + user + "/" + repository + "/master/" + project + "/Bravefile"

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		baseConfig = buf.String()
	}

	if len(baseConfig) == 0 {
		return nil, errors.New("Unable to download valid Bravefile. Check your URL")
	}

	err = yaml.Unmarshal([]byte(baseConfig), &bravefile)
	if err != nil {
		return nil, err
	}

	return &bravefile, nil
}

// GetBravefileFromLXD generates a Bravefile for import of images from LXD repository
func GetBravefileFromLXD(name string) (*Bravefile, error) {
	var bravefile Bravefile
	var baseConfig string

	dist := strings.SplitN(name, "/", -1)

	if len(dist) == 1 {
		return nil, errors.New("brave base accepts image names in the format NAME/VERSION/ARCH. See https://images.linuxcontainers.org for a list of supported images")
	}

	version := strings.SplitN(dist[1], ".", 2)
	distroVersion := version[0]

	if len(version) > 1 {
		distroVersion = strings.Join(version[:], "")
	}

	service := "brave-base-" + dist[0] + "-" + distroVersion

	baseConfig = BRAVEFILE

	nameRegexp, _ := regexp.Compile("<name>")
	serviceRegexp, _ := regexp.Compile("<service>")

	baseConfig = nameRegexp.ReplaceAllString(baseConfig, name)
	baseConfig = serviceRegexp.ReplaceAllString(baseConfig, service)

	err := yaml.Unmarshal([]byte(baseConfig), &bravefile)
	if err != nil {
		return nil, err
	}

	return &bravefile, nil
}
