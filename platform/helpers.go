package platform

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bravetools/bravetools/shared"
	"github.com/lxc/lxd/shared/api"
)

// Private Helpers

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
			return errors.New("Failed to mount to the source: " + err.Error())
		case "lxd":
			shared.ExecCommand(
				"lxc",
				"storage",
				"volume",
				"delete",
				storagePoolName,
				sharedDirectory)
			return errors.New("Failed to mount to the source: " + err.Error())
		}
	}

	// 3. Add storage volume as a disk device to target unit
	err = AddDevice(destUnit, sharedDirectory, shareSettings, bh.Remote)
	if err != nil {
		bh.UmountShare(sourceUnit, sharedDirectory)
		return errors.New("Failed to mount to the destination: " + err.Error())
	}

	return nil
}

func importLXD(bravefile *shared.Bravefile, remote Remote) (fingerprint string, err error) {
	fingerprint, err = Launch(bravefile.PlatformService.Name, bravefile.Base.Image, remote)
	if err != nil {
		return "", errors.New("Failed to launch base unit: " + err.Error())
	}

	return fingerprint, nil
}

func importGitHub(bravefile *shared.Bravefile, bh *BraveHost) error {
	home, _ := os.UserHomeDir()
	imageLocation := filepath.Join(home, shared.ImageStore)

	path := "github.com/" + bravefile.Base.Image
	remoteBravefile, err := shared.GetBravefileFromGitHub(path)
	if err != nil {
		return err
	}

	remoteServiceName := remoteBravefile.PlatformService.Name + "-" + remoteBravefile.PlatformService.Version

	if _, err := os.Stat(filepath.Join(imageLocation, remoteServiceName+".tar.gz")); os.IsNotExist(err) {
		err = bh.BuildImage(remoteBravefile)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Found local image " + remoteServiceName + ". Skipping GitHub build")
	}

	remoteBravefile.Base.Image = remoteServiceName
	remoteBravefile.PlatformService.Name = bravefile.PlatformService.Name

	err = importLocal(remoteBravefile, bh.Remote)
	if err != nil {
		return err
	}

	return nil
}

func importLocal(bravefile *shared.Bravefile, remote Remote) error {
	home, _ := os.UserHomeDir()
	location := filepath.Join(home, shared.ImageStore)

	fingerprint, err := ImportImage(filepath.Join(location, bravefile.Base.Image)+".tar.gz", bravefile.Base.Image, remote)

	if err != nil {
		return errors.New("Failed to import image: " + err.Error())
	}

	err = LaunchFromImage(bravefile.Base.Image, bravefile.PlatformService.Name, remote)
	if err != nil {
		DeleteImage(fingerprint, remote)
		return errors.New("Failed to launch unit: " + err.Error())
	}

	err = Start(bravefile.PlatformService.Name, remote)
	if err != nil {
		Delete(bravefile.PlatformService.Name, remote)
		return errors.New("Failed to start a unit: " + err.Error())
	}

	return nil
}

func copyTo(source string, settings HostSettings) error {

	backend := settings.BackendSettings.Type
	switch backend {
	case "multipass":
		err := shared.ExecCommand("multipass",
			"transfer",
			source,
			settings.BackendSettings.Resources.Name+":")
		if err != nil {
			return err
		}
	case "lxd":
		hd, _ := os.UserHomeDir()
		shared.CopyFile(source, hd)
	}

	return nil
}

// run script on host
func run(scriptPath string, settings HostSettings) error {

	backend := settings.BackendSettings.Type

	switch backend {
	case "multipass":
		err := shared.ExecCommand("multipass",
			"exec",
			settings.BackendSettings.Resources.Name,
			"--",
			"/bin/bash",
			scriptPath)
		if err != nil {
			return err
		}
	case "lxd":
		err := shared.ExecCommand(
			"sudo",
			"/bin/bash",
			scriptPath)
		if err != nil {
			return err
		}
	default:
		return errors.New("Cannot find backend")
	}

	return nil
}

func deleteHostImages(remote Remote) error {
	images, err := GetImages(remote)
	if err != nil {
		return errors.New("Failed to access host images: " + err.Error())
	}

	for _, i := range images {
		err := DeleteImage(i.Fingerprint, remote)
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

func getInterfaceName() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, errors.New("Failed to get network interfaces: " + err.Error())
	}

	var ifaceNames []string
	for _, i := range interfaces {
		addrs, _ := i.Addrs()
		name := i.Name

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				if !ip.IsLoopback() && ip.To4() != nil {
					addr := strings.Split(ip.String(), ".")
					if addr[3] != "1" {
						ifaceNames = append(ifaceNames, name)
					}
				}
			}
		}
	}

	return ifaceNames, err
}

func getMPInterfaceName(bh *BraveHost) ([]string, error) {

	grep := `ip -4 route ls | grep default | grep -Po '(?<=dev )(\S+)'`

	ifaceName, err := shared.ExecCommandWReturn(
		"multipass",
		"exec",
		bh.Settings.BackendSettings.Resources.Name,
		"--",
		"bash",
		"-c",
		grep)
	if err != nil {
		return nil, errors.New("Failed to get network interface name: " + err.Error())
	}

	ifaceName = strings.TrimRight(ifaceName, "\r\n")
	var ifaces []string
	ifaces = append(ifaces, ifaceName)

	return ifaces, nil
}

// ProcessInterruptHandler monitors for Ctrl+C keypress in Terminal
func processInterruptHandler(fingerprint string, bravefile *shared.Bravefile, bh *BraveHost) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Interrupting build and cleaning artefacts")
		DeleteImage(fingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)

		os.Exit(0)
	}()
}

func bravefileCopy(copy []shared.CopyCommand, service string, remote Remote) error {
	dir, _ := os.Getwd()
	for _, c := range copy {
		source := c.Source
		source = path.Join(dir, source)
		sourcePath := filepath.FromSlash(source)

		target := c.Target
		_, err := Exec(service, []string{"mkdir", "-p", target}, remote)
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
			_, err = Exec(service, []string{"bash", "-c", c.Action}, remote)
			if err != nil {
				return errors.New("Failed to execute action: " + err.Error())
			}
		}
	}

	return nil
}

func bravefileRun(run []shared.RunCommand, service string, remote Remote) (status int, err error) {
	for _, c := range run {
		var command string
		var content string

		if c.Command != "" {
			command = c.Command
		}

		args := []string{command}
		if len(c.Args) > 0 {
			for _, a := range c.Args {
				args = append(args, a)
			}
		}
		if c.Content != "" {
			content = c.Content
			args = append(args, content)
		}

		status, err = Exec(service, args, remote)

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

func getImageFingerprint(slice1 []api.Image, slice2 []api.Image) string {
	var diff []string
	var fingerprint string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1.Fingerprint == s2.Fingerprint {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1.Fingerprint)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	if len(diff) > 0 {
		fingerprint = diff[0]
	}

	return fingerprint
}
