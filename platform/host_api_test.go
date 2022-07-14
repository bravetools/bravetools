package platform

import (
	"context"
	"testing"

	"github.com/bravetools/bravetools/shared"
)

func Test_DeleteLocalImage(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

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
	bravefile.PlatformService.Version = "1.0"

	err = host.BuildImage(&bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.DeleteLocalImage("alpine-test-1.0")
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}
}

func Test_InitUnit(t *testing.T) {
	host, err := NewBraveHost()
	if err != nil {
		t.Fatal("failed to create host: ", err.Error())
	}

	ctx := context.Background()

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

	err = host.BuildImage(&bravefile)
	if err != nil {
		t.Error("host.BuildImage: ", err)
	}

	err = host.InitUnit(host.Backend, &bravefile.PlatformService)
	if err != nil {
		t.Error("host.InitUnit: ", err)
	}

	err = host.Postdeploy(ctx, &bravefile.PlatformService)
	if err != nil {
		t.Error("host.Postdeploy: ", err)
	}

	err = host.DeleteLocalImage("alpine-test-1.0")
	if err != nil {
		t.Error("host.DeleteImageByName: ", err)
	}

	err = host.StopUnit("alpine-test", host.Backend)
	if err != nil {
		t.Error("host.StopUnit: ", err)
	}

	err = host.StartUnit("alpine-test", host.Backend)
	if err != nil {
		t.Error("host.StartUnit: ", err)
	}

	err = host.DeleteUnit("alpine-test")
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

	err = host.ListLocalImages()
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

	err = host.ListUnits(host.Backend)
	if err != nil {
		t.Error("host.ListLocalImages: ", err)
	}
}
