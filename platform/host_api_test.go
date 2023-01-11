package platform

import (
	"log"
	"testing"

	"github.com/bravetools/bravetools/shared"
)

func Test_DeleteLocalImage(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	imageName := "alpine/edge/amd64"

	bravefile, err := GetBravefileFromLXD(imageName)
	if err != nil {
		t.Error("platform.GetBravefileFromLXD: ", err)
	}

	err = host.BuildImage(*bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.DeleteLocalImage(imageName, false)
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}

}

func Test_HostInfo(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	err = host.HostInfo(false)
	if err != nil {
		t.Error("host.HostInfo: ", err)
	}
}

func Test_BuildImage(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	bravefile := *shared.NewBravefile()
	bravefile.Base.Image = "alpine/edge/amd64"
	bravefile.Base.Location = "public"

	bravefile.SystemPackages.Manager = "apk"
	bravefile.SystemPackages.System = []string{"htop", "make"}

	runCommand := &shared.RunCommand{}
	runCommand.Command = "echo"
	runCommand.Args = []string{"Hello World"}

	bravefile.Run = []shared.RunCommand{*runCommand}

	bravefile.PlatformService.Name = "alpine-test"
	bravefile.PlatformService.Image = "alpine-test-1.0"
	bravefile.PlatformService.Version = "1.0"

	err = host.BuildImage(bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.DeleteLocalImage(bravefile.PlatformService.Image, true)
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}
}

func Test_InitUnit(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	bravefile := *shared.NewBravefile()
	bravefile.Base.Image = "alpine/edge/amd64"
	bravefile.Base.Location = "public"

	bravefile.SystemPackages.Manager = "apk"
	bravefile.SystemPackages.System = []string{"htop", "make"}

	runCommand := &shared.RunCommand{}
	runCommand.Command = "echo"
	runCommand.Args = []string{"Hello World"}

	bravefile.Run = []shared.RunCommand{*runCommand}

	bravefile.PlatformService.Name = "alpine-test"
	bravefile.PlatformService.Version = "1.0"
	bravefile.PlatformService.Image = "alpine-test-1.0"

	bravefile.PlatformService.Resources.CPU = "1"
	bravefile.PlatformService.Resources.RAM = "1GB"

	bravefile.PlatformService.Postdeploy.Run = []shared.RunCommand{
		{
			Command: "echo",
			Args:    []string{"Hello World"},
		},
	}

	err = host.BuildImage(bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.InitUnit(host.Backend, bravefile.PlatformService)
	if err != nil {
		t.Error("host.InitUnit: ", err)
	}

	err = host.DeleteLocalImage(bravefile.PlatformService.Image, true)
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}

	err = host.StopUnit(bravefile.PlatformService.Name)
	if err != nil {
		t.Error("host.StopUnit: ", err)
	}

	err = host.StartUnit(bravefile.PlatformService.Name)
	if err != nil {
		t.Error("host.StartUnit: ", err)
	}

	err = host.DeleteUnit(bravefile.PlatformService.Name)
	if err != nil {
		t.Error("host.DeleteUnit: ", err)
	}
}

func Test_ListLocalImages(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	err = host.HostInfo(false)
	if err != nil {
		t.Error("host.HostInfo: ", err)
	}

	err = host.PrintLocalImages()
	if err != nil {
		t.Error("host.ListLocalImages: ", err)
	}
}

func Test_ListUnits(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	err = host.HostInfo(false)
	if err != nil {
		t.Error("host.HostInfo: ", err)
	}

	err = host.PrintUnits(host.Backend, "")
	if err != nil {
		t.Error("host.ListLocalImages: ", err)
	}
}

func Test_Compose(t *testing.T) {
	var err error

	host, err := NewBraveHost()
	if err != nil {
		log.Fatal(err)
	}

	composefile := shared.NewComposeFile()
	err = composefile.Load("../test/compose/python-multi-service/brave-compose.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// Composefile specifies static IP addresses for containers, test will fail on bravetools setups where LXD bridge is on different IP range
	// For this test, will just clear static IP addresses
	for _, service := range composefile.Services {
		service.IP = ""
	}

	err = host.Compose(host.Backend, composefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	for _, service := range composefile.Services {
		err = host.DeleteUnit(service.Name)
		if err != nil {
			t.Errorf("failed to delete unit: %q", service.Name)
			t.Log(err)
		}
		err = host.DeleteLocalImage(service.Image, true)
		if err != nil {
			t.Errorf("failed to delete unit: %q", service.Image)
			t.Log(err)
		}
	}
}
