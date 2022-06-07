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

	"github.com/bravetools/bravetools/shared"
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

func createSharedVolume(storagePoolName string,
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
	err := AddDevice(sourceUnit, sharedDirectory, shareSettings, bh.Remote)
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
	err = AddDevice(destUnit, sharedDirectory, shareSettings, bh.Remote)
	if err != nil {
		bh.UmountShare(sourceUnit, sharedDirectory)
		return errors.New("failed to mount to destination: " + err.Error())
	}

	return nil
}

func importLXD(ctx context.Context, bravefile *shared.Bravefile, remote Remote) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	fingerprint, err = Launch(ctx, bravefile.PlatformService.Name, bravefile.Base.Image, remote)
	if err != nil {
		return fingerprint, errors.New("failed to launch base unit: " + err.Error())
	}

	return fingerprint, nil
}

func importGitHub(ctx context.Context, bravefile *shared.Bravefile, bh *BraveHost) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	home, _ := os.UserHomeDir()
	imageLocation := filepath.Join(home, shared.ImageStore)

	path := "github.com/" + bravefile.Base.Image
	remoteBravefile, err := shared.GetBravefileFromGitHub(path)
	if err != nil {
		return fingerprint, err
	}

	remoteServiceName := remoteBravefile.PlatformService.Name + "-" + remoteBravefile.PlatformService.Version

	if _, err := os.Stat(filepath.Join(imageLocation, remoteServiceName+".tar.gz")); os.IsNotExist(err) {
		err = bh.BuildImage(remoteBravefile)
		if err != nil {
			return fingerprint, err
		}
	} else {
		fmt.Println("Found local image " + remoteServiceName + ". Skipping GitHub build")
	}

	remoteBravefile.Base.Image = remoteServiceName
	remoteBravefile.PlatformService.Name = bravefile.PlatformService.Name

	fingerprint, err = importLocal(ctx, remoteBravefile, bh.Remote)
	return fingerprint, err
}

func importLocal(ctx context.Context, bravefile *shared.Bravefile, remote Remote) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	home, _ := os.UserHomeDir()
	location := filepath.Join(home, shared.ImageStore)

	fingerprint, err = ImportImage(filepath.Join(location, bravefile.Base.Image)+".tar.gz", bravefile.Base.Image, remote)

	if err != nil {
		return fingerprint, errors.New("failed to import image: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	err = LaunchFromImage(bravefile.Base.Image, bravefile.PlatformService.Name, remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, remote)
		return fingerprint, errors.New("failed to launch unit: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	err = Start(bravefile.PlatformService.Name, remote)
	if err != nil {
		DeleteUnit(bravefile.PlatformService.Name, remote)
		DeleteImageByFingerprint(fingerprint, remote)
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

func deleteHostImages(remote Remote) error {
	images, err := GetImages(remote)
	if err != nil {
		return errors.New("Failed to access host images: " + err.Error())
	}

	for _, i := range images {
		err := DeleteImageByFingerprint(i.Fingerprint, remote)
		if err != nil {
			return errors.New("Failed to delete image: " + i.Fingerprint)
		}
	}

	return nil
}

func listHostImages(remote Remote) ([]api.Image, error) {
	images, err := GetImages(remote)
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

func cleanupBuild(imageFingerprint string, bravefile *shared.Bravefile, bh *BraveHost) {
	DeleteUnit(bravefile.PlatformService.Name, bh.Remote)
	DeleteImageByFingerprint(imageFingerprint, bh.Remote)
}

func bravefileCopy(ctx context.Context, copy []shared.CopyCommand, service string, remote Remote) error {
	dir, _ := os.Getwd()
	for _, c := range copy {
		if err := ctx.Err(); err != nil {
			return err
		}

		source := c.Source
		source = path.Join(dir, source)
		sourcePath := filepath.FromSlash(source)

		target := c.Target
		_, err := Exec(ctx, service, []string{"mkdir", "-p", target}, remote)
		if err != nil {
			return errors.New("Failed to create target directory: " + err.Error())
		}

		fi, err := os.Lstat(sourcePath)
		if err != nil {
			return errors.New("Failed to read file " + sourcePath + ": " + err.Error())
		}

		if fi.IsDir() {
			err = Push(service, sourcePath, target, remote)
			if err != nil {
				return errors.New("Failed to push symlink: " + err.Error())
			}
		} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			err = SymlinkPush(service, sourcePath, target, remote)
			if err != nil {
				return errors.New("Failed to push directory: " + err.Error())
			}
		} else {
			err = FilePush(service, sourcePath, target, remote)
			if err != nil {
				return errors.New("Failed to push file: " + err.Error())
			}
		}

		if c.Action != "" {
			_, err = Exec(ctx, service, []string{"bash", "-c", c.Action}, remote)
			if err != nil {
				return errors.New("Failed to execute action: " + err.Error())
			}
		}
	}

	return nil
}

func bravefileRun(ctx context.Context, run []shared.RunCommand, service string, remote Remote) (status int, err error) {
	for _, c := range run {
		if err = ctx.Err(); err != nil {
			return 1, err
		}

		var command string
		var content string

		if c.Command != "" {
			command = c.Command
		}

		args := []string{command}
		if len(c.Args) > 0 {
			args = append(args, c.Args...)
			// for _, a := range c.Args {
			// 	args = append(args, a)
			// }
		}
		if c.Content != "" {
			content = c.Content
			args = append(args, content)
		}

		status, err = Exec(ctx, service, args, remote)

	}

	return status, err
}

func cleanUnusedStoragePool(name string, remote Remote) {
	err := DeleteStoragePool(name, remote)
	if err != nil {
		fmt.Println("Nothing to clean")
	}
}

// addIPRules adds firewall rule to the host iptable
func addIPRules(ct string, hostPort string, ctPort string, bh *BraveHost) error {

	name := ct + "-proxy-" + hostPort + "-" + ctPort

	var config = make(map[string]string)

	config["type"] = "proxy"
	config["listen"] = "tcp:0.0.0.0:" + hostPort
	config["connect"] = "tcp:127.0.0.1:" + ctPort

	err := AddDevice(ct, name, config, bh.Remote)
	if err != nil {
		return errors.New("failed to add proxy settings for unit " + err.Error())
	}

	return nil
}

func checkUnits(unitName string, bh *BraveHost) error {
	// Unit Checks
	unitList, err := GetUnits(bh.Remote)
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
