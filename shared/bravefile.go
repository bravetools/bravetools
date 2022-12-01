package shared

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

// ImageDescription defines base image type and source
type ImageDescription struct {
	Image        string `yaml:"image"`
	Location     string `yaml:"location"`
	Architecture string `yaml:"architecture"`
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
	Detach  bool              `yaml:"detach,omitempty"`
}

// CopyCommand defines source and target for files to be copied into container
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
	Profile    string     `yaml:"profile,omitempty"`
	Storage    string     `yaml:"storage,omitempty"`
	Network    string     `yaml:"network,omitempty"`
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
	Image           string           `yaml:"image,omitempty"`
	Base            ImageDescription `yaml:"base,omitempty"`
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

	if bravefile.Image != "" && bravefile.PlatformService.Version != "" {
		return fmt.Errorf("bravefile at path %q uses legacy 'version' field in service section and 'image' field in build section "+
			"- define version in 'image' field using <image_name>[/version][/arch]", file)
	}

	if bravefile.Image == "" && bravefile.PlatformService.Image == "" {
		return fmt.Errorf("image not defined in bravefile")
	}

	if bravefile.Image != "" && bravefile.PlatformService.Image != "" {
		if bravefile.Image != bravefile.PlatformService.Image {
			return fmt.Errorf("two different images defined in same Bravfile: %q and %q", bravefile.Image, bravefile.PlatformService.Image)
		}
	}

	return nil
}

// Validate validates Bravefile for build
func (bravefile *Bravefile) ValidateBuild() error {
	if bravefile.Base.Image == "" {
		return errors.New("invalid Bravefile: empty Base Image name")
	}

	if bravefile.Image == "" && bravefile.PlatformService.Image == "" {
		return errors.New("invalid Bravefile: empty Service Image name")
	}

	return nil
}

func (bravefile *Bravefile) IsLegacy() bool {
	if bravefile.PlatformService.Version == "" || bravefile.Image != "" {
		return false
	}

	return true
}

func (service *Service) IsLegacy() bool {
	if service.Version == "" {
		return false
	}

	return true
}

func (service *Service) ValidateDeploy() error {
	if service.Name == "" {
		return errors.New("invalid Service: empty Service Name")
	}

	if service.Image == "" {
		return fmt.Errorf("invalid Service %q: empty Image name", service.Name)
	}

	if strings.ContainsAny(service.Name, "/_. !@£$%^&*(){};`~,?") {
		return errors.New("unit names should not contain special characters")
	}

	if len(service.Ports) > 0 {
		for _, p := range service.Ports {
			ps := strings.Split(p, ":")
			if len(ps) < 2 || len(ps) > 2 || ps[0] == "" || ps[1] == "" {
				return fmt.Errorf("invalid port forwarding definition %q. Appropriate format is UNIT_PORT:HOST_PORT", p)
			}
		}
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
	if s.Profile == "" {
		s.Profile = service.Profile
	}
	if s.Network == "" {
		s.Network = service.Network
	}
	if s.Storage == "" {
		s.Storage = service.Storage
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
		return nil, errors.New("unable to download valid Bravefile. Check your URL")
	}

	err = yaml.Unmarshal([]byte(baseConfig), &bravefile)
	if err != nil {
		return nil, err
	}

	return &bravefile, nil
}
