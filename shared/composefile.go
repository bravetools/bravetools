package shared

import "gopkg.in/yaml.v2"

type ComposeService struct {
	Service   `yaml:",inline"`
	Bravefile string   `yaml:"bravefile,omitempty"`
	Build     bool     `yaml:"build,omitempty"`
	Depends   []string `yaml:"depends_on,omitempty"`
}

type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
}

func NewComposeFile() *ComposeFile {
	return &ComposeFile{}
}

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
