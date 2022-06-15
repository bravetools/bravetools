package platform

import (
	"testing"

	"github.com/bravetools/bravetools/shared"
)

func Test_DeleteLocalImage(t *testing.T) {
	host := *NewBraveHost()

	bravefile, err := shared.GetBravefileFromLXD("alpine/edge/amd64")
	if err != nil {
		t.Error("shared.GetBravefileFromLXD: ", err)
	}

	err = host.BuildImage(bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.DeleteLocalImage("brave-base-alpine-edge-1.0")
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}

}

func Test_HostInfo(t *testing.T) {
	host := *NewBraveHost()
	backend := NewLxd(host.Settings)

	err := host.HostInfo(backend, false)
	if err != nil {
		t.Error("host.HostInfo: ", err)
	}
}

func Test_BuildImage(t *testing.T) {
	host := *NewBraveHost()

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

	err := host.BuildImage(&bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.DeleteLocalImage("alpine-test-1.0")
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}
}

func Test_InitUnit(t *testing.T) {
	host := *NewBraveHost()
	backend := NewLxd(host.Settings)

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

	err := host.BuildImage(&bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.InitUnit(backend, &bravefile)
	if err != nil {
		t.Error("host.InitUnit: ", err)
	}

	err = host.DeleteLocalImage("alpine-test-1.0")
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}
}
