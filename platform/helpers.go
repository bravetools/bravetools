package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bravetools/bravetools/shared"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

// Private Helpers

func getCurrentUsername() (username string, err error) {

	user, err := user.Current()
	if err != nil {
		return "", err
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}

	username = reg.ReplaceAllString(user.Username, "")

	// Truncate to max username length if necessary
	usernameLength := 12
	if len(username) < usernameLength {
		usernameLength = len(username)
	}
	username = username[:usernameLength]

	return username, nil
}

func createSharedVolume(lxdServer lxd.InstanceServer,
	storagePoolName string,
	sharedDirectory string,
	sourceUnit string,
	destUnit string,
	destPath string,
	bh *BraveHost) error {

	backend := bh.Settings.BackendSettings.Type

	switch backend {
	case "multipass":
		// 1. Create storage volume
		err := shared.ExecCommand(
			"multipass",
			"exec",
			bh.Settings.BackendSettings.Resources.Name,
			"--",
			shared.SnapLXC,
			"storage",
			"volume",
			"create",
			storagePoolName,
			sharedDirectory)
		if err != nil {
			return errors.New("Failed to create storage volume: " + sharedDirectory + ": " + err.Error())
		}
	case "lxd":
		// 1. Create storage volume
		err := shared.ExecCommand(
			"lxc",
			"storage",
			"volume",
			"create",
			storagePoolName,
			sharedDirectory)
		if err != nil {
			return errors.New("Failed to create storage volume: " + sharedDirectory + ": " + err.Error())
		}
	}

	shareSettings := map[string]string{}
	shareSettings["path"] = destPath
	shareSettings["pool"] = storagePoolName
	shareSettings["source"] = sharedDirectory
	shareSettings["type"] = "disk"

	// 2. Add storage volume as a disk device to source unit
	err := AddDevice(lxdServer, sourceUnit, sharedDirectory, shareSettings)
	if err != nil {
		switch backend {
		case "multipass":
			shared.ExecCommand(
				"multipass",
				"exec",
				bh.Settings.BackendSettings.Resources.Name,
				"--",
				shared.SnapLXC,
				"lxc",
				"storage",
				"volume",
				"delete",
				storagePoolName,
				sharedDirectory)
			return errors.New("Failed to mount to source: " + err.Error())
		case "lxd":
			shared.ExecCommand(
				"lxc",
				"storage",
				"volume",
				"delete",
				storagePoolName,
				sharedDirectory)
			return errors.New("failed to mount to source: " + err.Error())
		}
	}

	// 3. Add storage volume as a disk device to target unit
	err = AddDevice(lxdServer, destUnit, sharedDirectory, shareSettings)
	if err != nil {
		bh.UmountShare(sourceUnit, sharedDirectory)
		return errors.New("failed to mount to destination: " + err.Error())
	}

	return nil
}

func importLXD(ctx context.Context, lxdServer lxd.InstanceServer, bravefile *shared.Bravefile, profileName string) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	fingerprint, err = Launch(ctx, lxdServer, bravefile.PlatformService.Name, bravefile.Base.Image, profileName)
	if err != nil {
		return fingerprint, errors.New("failed to launch base unit: " + err.Error())
	}

	return fingerprint, nil
}

func importGitHub(ctx context.Context, lxdServer lxd.InstanceServer, bravefile *shared.Bravefile, bh *BraveHost, profileName string, storagePool string) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}

	path := bravefile.Base.Image
	if !strings.HasPrefix(path, "github.com/") {
		path = "github.com/" + path
	}
	remoteBravefile, err := shared.GetBravefileFromGitHub(path)
	if err != nil {
		return fingerprint, err
	}

	var imageStruct BravetoolsImage

	// If version explicitly provided separately this is a legacy Bravefile
	if remoteBravefile.PlatformService.Version == "" {
		imageStruct, err = ParseImageString(remoteBravefile.PlatformService.Image)
	} else {
		imageStruct, err = ParseLegacyImageString(remoteBravefile.PlatformService.Image)
	}
	if err != nil {
		return fingerprint, err
	}
	// Ensure that the up-to-date "version" value is in the Bravefile for later use
	remoteBravefile.PlatformService.Version = imageStruct.Version

	if !imageExists(imageStruct) {
		err = bh.BuildImage(remoteBravefile)
		if err != nil {
			return fingerprint, err
		}
	} else {
		fmt.Println("Found local image " + imageStruct.String() + ". Skipping GitHub build")
	}

	remoteBravefile.Base.Image = imageStruct.String()
	remoteBravefile.PlatformService.Name = bravefile.PlatformService.Name
	// Since we are using new image format above, we need to set version to "" to prevent parsing as legacy image name
	remoteBravefile.PlatformService.Version = ""

	fingerprint, err = importLocal(ctx, lxdServer, remoteBravefile, profileName, storagePool)
	return fingerprint, err
}

func importLocal(ctx context.Context, lxdServer lxd.InstanceServer, bravefile *shared.Bravefile, profileName string, storagePool string) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	var imageStruct BravetoolsImage

	// If version explicitly provided separately this is a legacy Bravefile
	if bravefile.PlatformService.Version == "" {
		imageStruct, err = ParseImageString(bravefile.Base.Image)
	} else {
		imageStruct, err = ParseLegacyImageString(bravefile.Base.Image)
	}
	if err != nil {
		return "", err
	}
	// Ensure that the up-to-date "version" value is in the Bravefile for later use
	bravefile.PlatformService.Version = imageStruct.Version

	path, err := getImageFilepath(imageStruct)
	if err != nil {
		return "", err
	}

	fingerprint, err = shared.FileSha256Hash(path)
	if err != nil {
		return fingerprint, err
	}

	_, err = ImportImage(lxdServer, path, bravefile.Base.Image)
	if err != nil {
		return fingerprint, errors.New("failed to import image: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	err = LaunchFromImage(lxdServer, bravefile.Base.Image, bravefile.PlatformService.Name, profileName, storagePool)
	if err != nil {
		return fingerprint, errors.New("failed to launch unit: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	err = Start(lxdServer, bravefile.PlatformService.Name)
	if err != nil {
		return fingerprint, errors.New("failed to start a unit: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	return fingerprint, nil
}

// func copyTo(source string, settings HostSettings) error {

// 	backend := settings.BackendSettings.Type
// 	switch backend {
// 	case "multipass":
// 		err := shared.ExecCommand("multipass",
// 			"transfer",
// 			source,
// 			settings.BackendSettings.Resources.Name+":")
// 		if err != nil {
// 			return err
// 		}
// 	case "lxd":
// 		hd, _ := os.UserHomeDir()
// 		shared.CopyFile(source, hd)
// 	}

// 	return nil
// }

// // run script on host
// func run(scriptPath string, settings HostSettings) error {

// 	backend := settings.BackendSettings.Type

// 	switch backend {
// 	case "multipass":
// 		err := shared.ExecCommand("multipass",
// 			"exec",
// 			settings.BackendSettings.Resources.Name,
// 			"--",
// 			"/bin/bash",
// 			scriptPath)
// 		if err != nil {
// 			return err
// 		}
// 	case "lxd":
// 		err := shared.ExecCommand(
// 			"sudo",
// 			"/bin/bash",
// 			scriptPath)
// 		if err != nil {
// 			return err
// 		}
// 	default:
// 		return errors.New("cannot find backend")
// 	}

// 	return nil
// }

func deleteHostImages(lxdServer lxd.InstanceServer) error {
	images, err := GetImages(lxdServer)
	if err != nil {
		return errors.New("Failed to access host images: " + err.Error())
	}

	for _, i := range images {
		err := DeleteImageByFingerprint(lxdServer, i.Fingerprint)
		if err != nil {
			return errors.New("Failed to delete image: " + i.Fingerprint)
		}
	}

	return nil
}

func listHostImages(lxdServer lxd.ImageServer) ([]api.Image, error) {
	images, err := GetImages(lxdServer)
	if err != nil {
		return nil, errors.New("Failed to access host images: " + err.Error())
	}

	return images, nil
}

// func getInterfaceName() ([]string, error) {
// 	interfaces, err := net.Interfaces()
// 	if err != nil {
// 		return nil, errors.New("failed to get network interfaces: " + err.Error())
// 	}

// 	var ifaceNames []string
// 	for _, i := range interfaces {
// 		addrs, _ := i.Addrs()
// 		name := i.Name

// 		for _, addr := range addrs {
// 			var ip net.IP
// 			switch v := addr.(type) {
// 			case *net.IPNet:
// 				ip = v.IP
// 				if !ip.IsLoopback() && ip.To4() != nil {
// 					addr := strings.Split(ip.String(), ".")
// 					if addr[3] != "1" {
// 						ifaceNames = append(ifaceNames, name)
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return ifaceNames, err
// }

// func getMPInterfaceName(bh *BraveHost) ([]string, error) {

// 	grep := `ip -4 route ls | grep default | grep -Po '(?<=dev )(\S+)'`

// 	ifaceName, err := shared.ExecCommandWReturn(
// 		"multipass",
// 		"exec",
// 		bh.Settings.BackendSettings.Resources.Name,
// 		"--",
// 		"bash",
// 		"-c",
// 		grep)
// 	if err != nil {
// 		return nil, errors.New("failed to get network interface name: " + err.Error())
// 	}

// 	ifaceName = strings.TrimRight(ifaceName, "\r\n")
// 	var ifaces []string
// 	ifaces = append(ifaces, ifaceName)

// 	return ifaces, nil
// }

// postdeploy copy files and run commands on running service
func postdeploy(ctx context.Context, lxdServer lxd.InstanceServer, unitConfig *shared.Service) (err error) {

	if unitConfig.Postdeploy.Copy != nil {
		err = bravefileCopy(ctx, lxdServer, unitConfig.Postdeploy.Copy, unitConfig.Name)
		if err != nil {
			return err
		}
	}

	if unitConfig.Postdeploy.Run != nil {
		err = bravefileRun(ctx, lxdServer, unitConfig.Postdeploy.Run, unitConfig.Name)
		if err != nil {
			return errors.New(shared.Fatal("failed to execute command: " + err.Error()))
		}
	}

	return nil
}

func bravefileCopy(ctx context.Context, lxdServer lxd.InstanceServer, copy []shared.CopyCommand, service string) error {
	dir, _ := os.Getwd()
	for _, c := range copy {
		if err := ctx.Err(); err != nil {
			return err
		}

		source := c.Source
		source = path.Join(dir, source)
		sourcePath := filepath.FromSlash(source)

		target := c.Target
		_, err := Exec(ctx, lxdServer, service, []string{"mkdir", "-p", target}, ExecArgs{})
		if err != nil {
			return errors.New("Failed to create target directory: " + err.Error())
		}

		fi, err := os.Lstat(sourcePath)
		if err != nil {
			return errors.New("Failed to read file " + sourcePath + ": " + err.Error())
		}

		if fi.IsDir() {
			err = Push(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push symlink: " + err.Error())
			}
		} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			err = SymlinkPush(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push directory: " + err.Error())
			}
		} else {
			err = FilePush(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push file: " + err.Error())
			}
		}

		if c.Action != "" {
			_, err = Exec(ctx, lxdServer, service, []string{"bash", "-c", c.Action}, ExecArgs{})
			if err != nil {
				return errors.New("Failed to execute action: " + err.Error())
			}
		}
	}

	return nil
}

func bravefileRun(ctx context.Context, lxdServer lxd.InstanceServer, run []shared.RunCommand, service string) (err error) {
	for _, c := range run {
		if err = ctx.Err(); err != nil {
			return err
		}

		var command string
		var content string

		if c.Command != "" {
			command = c.Command
		}

		args := []string{command}
		if len(c.Args) > 0 {
			args = append(args, c.Args...)
		}
		if c.Content != "" {
			content = c.Content
			args = append(args, content)
		}

		status, err := Exec(ctx, lxdServer, service, args, ExecArgs{env: c.Env})
		if err != nil {
			return err
		}
		if status > 0 {
			return fmt.Errorf("non-zero exit code %d for command %q", status, strings.Join(args, " "))
		}
	}

	return err
}

func cleanUnusedStoragePool(lxdServer lxd.InstanceServer, name string) {
	err := DeleteStoragePool(lxdServer, name)
	if err != nil {
		fmt.Println("Nothing to clean")
	}
}

// addIPRules adds firewall rule to the host iptable

func addIPRules(lxdServer lxd.InstanceServer, ct string, hostPort string, ctPort string) error {

	name := ct + "-proxy-" + hostPort + "-" + ctPort

	var config = make(map[string]string)

	config["type"] = "proxy"
	config["listen"] = "tcp:0.0.0.0:" + hostPort
	config["connect"] = "tcp:127.0.0.1:" + ctPort

	err := AddDevice(lxdServer, ct, name, config)
	if err != nil {
		return errors.New("failed to add proxy settings for unit " + err.Error())
	}

	return nil
}

func checkUnits(lxdServer lxd.InstanceServer, unitName string, profileName string) error {
	if unitName == "" {
		return errors.New("unit name cannot be empty")
	}

	// Unit Checks
	unitList, err := GetUnits(lxdServer, profileName)
	if err != nil {
		return err
	}

	var unitNames []string
	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	unitExists := shared.StringInSlice(unitName, unitNames)
	if unitExists {
		return errors.New("Unit " + unitName + " already exists on host")
	}

	return nil
}

func getBaseOnlyServices(composeFile *shared.ComposeFile) (serviceNames []string) {
	for serviceName := range composeFile.Services {
		if composeFile.Services[serviceName].Base && !composeFile.Services[serviceName].Build {
			serviceNames = append(serviceNames, serviceName)
		}
	}
	return serviceNames
}

func getBuildDependents(dependency string, composeFile *shared.ComposeFile) (serviceNames []string, err error) {
	for service := range composeFile.Services {
		var imageStruct BravetoolsImage

		// If version explicitly provided separately this is a legacy Bravefile
		if composeFile.Services[service].Version == "" {
			imageStruct, err = ParseImageString(composeFile.Services[service].Image)
		} else {
			imageStruct, err = ParseLegacyImageString(composeFile.Services[service].Image)
		}
		if err != nil {
			return serviceNames, err
		}
		// Ensure that the up-to-date "version" value is in the compose file for later use
		composeFile.Services[service].Version = imageStruct.Version

		if imageExists(imageStruct) {
			continue
		}
		for _, dependsOn := range composeFile.Services[service].Depends {
			if dependsOn == dependency {
				serviceNames = append(serviceNames, service)
			}
		}
	}
	return serviceNames, nil
}
