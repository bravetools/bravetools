package shared

import "testing"

func TestValidateDeployPorts(t *testing.T) {
	service := Service{Name: "test-container", Image: "test-image", Ports: []string{"3000"}}
	if err := service.ValidateDeploy(); err == nil {
		t.Errorf("Expected port forwarding %q to fail", service.Ports)
	}

	service.Ports = []string{"3000:3000:3000"}
	if err := service.ValidateDeploy(); err == nil {
		t.Errorf("Expected port forwarding %q to fail", service.Ports)
	}

	service.Ports = []string{"3000:"}
	if err := service.ValidateDeploy(); err == nil {
		t.Errorf("Expected port forwarding %q to fail", service.Ports)
	}

	service.Ports = []string{"3000:3000"}
	if err := service.ValidateDeploy(); err != nil {
		t.Errorf("Expected port forwarding %q to succeed", service.Ports)
	}

	service.Ports = []string{}
	if err := service.ValidateDeploy(); err != nil {
		t.Errorf("Expected empty port forwarding %q to succeed", service.Ports)
	}
}
