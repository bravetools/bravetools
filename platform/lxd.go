package platform

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bravetools/bravetools/shared"
)

type (
	// Lxd ..
	Lxd struct {
		Settings *HostSettings
	}
)

// LxdStatus enum
type LxdStatus int

// LxdStatus constants
const (
	NotInstalled LxdStatus = iota
	NotInitialised
	Incompatible
	Installed
)

func deleteBraveHome() error {
	userHome, _ := os.UserHomeDir()
	bravetoolsHome := path.Join(userHome, shared.BraveHome)
	err := os.RemoveAll(bravetoolsHome)
	return err
}

func lxdCheck(vm Lxd) (status LxdStatus, lxcPath string, err error) {
	switch vm.Settings.BackendSettings.Type {
	case "lxd":
		lxcPath, err = shared.ExecCommandWReturn(
			"which",
			"lxc")
	case "multipass":
		lxcPath, err = shared.ExecCommandWReturn(
			"multipass",
			"exec",
			vm.Settings.Name,
			"--",
			"which",
			"lxc")
	}
	if err != nil {
		return -1, "", err
	}

	if lxcPath == "" {
		return NotInstalled, "", nil
	}

	lxcPath = strings.TrimSpace(lxcPath)

	if len(lxcPath) > 0 && vm.Settings.Status == "inactive" {
		return NotInitialised, lxcPath, nil
	}

	return Installed, lxcPath, nil
}

// NewLxd constructor
func NewLxd(settings HostSettings) *Lxd {
	return &Lxd{
		Settings: &settings,
	}
}

// BraveBackendInit ..
func (vm Lxd) BraveBackendInit() error {
	lxdStatus, whichLxc, err := lxdCheck(vm)
	if err != nil {
		return errors.New("failed to identify LXD: " + err.Error())
	}

	switch lxdStatus {
	case Incompatible:
		_ = deleteBraveHome()
		return errors.New("incompatible LXD version")
	case NotInstalled:
		_ = deleteBraveHome()
		return errors.New("LXD not installed")

	case NotInitialised:
		err = initiateLxd(vm, whichLxc)
		if err != nil {
			_ = deleteBraveHome()
			return errors.New("failed to initiate LXD: " + err.Error())
		}

		err = enableRemote(vm, whichLxc)
		if err != nil {
			_ = deleteBraveHome()
			return errors.New("failed to enable remote: " + err.Error())
		}

		return nil
	case Installed:
		return errors.New("bravetools is already initialised. Run \"brave configure\" if you'd like to tweak configuration")

	default:
		return nil
	}
}

func initiateLxd(vm Lxd, whichLxc string) error {

	_, _, err := checkLXDVersion(whichLxc)
	if err != nil {
		return err
	}

	err = shared.ExecCommand(
		whichLxc,
		"profile",
		"create",
		vm.Settings.Profile)
	if err != nil {
		return errors.New("Failed to create LXD profile: " + err.Error())
	}

	err = shared.ExecCommand(
		whichLxc,
		"storage",
		"create",
		vm.Settings.StoragePool.Name,
		vm.Settings.StoragePool.Type,
		"size="+vm.Settings.StoragePool.Size)
	if err != nil {
		_ = shared.ExecCommand(whichLxc, "profile", "delete", vm.Settings.Profile)
		return errors.New("Failed to create storage pool: " + err.Error())
	}

	bridge := "ipv4.address=" + vm.Settings.Network.Bridge + "/24"

	err = shared.ExecCommand(
		whichLxc,
		"network",
		"create",
		vm.Settings.Profile+"br0",
		"ipv6.address=none",
		bridge,
		"ipv4.nat=true")
	if err != nil {
		_ = shared.ExecCommand(whichLxc, "profile", "delete", vm.Settings.Profile)
		_ = shared.ExecCommand(whichLxc, "storage", "delete", vm.Settings.StoragePool.Name)

		return errors.New("Failed to create network: " + err.Error())
	}

	err = shared.ExecCommand(
		whichLxc,
		"network",
		"attach-profile",
		vm.Settings.Profile+"br0",
		vm.Settings.Profile,
		"eth0")
	if err != nil {
		_ = shared.ExecCommand(whichLxc, "profile", "delete", vm.Settings.Profile)
		_ = shared.ExecCommand(whichLxc, "storage", "delete", vm.Settings.StoragePool.Name)
		_ = shared.ExecCommand(whichLxc, "network", "delete", vm.Settings.Profile+"br0")

		return errors.New("Failed to attach network to profile: " + err.Error())
	}

	shared.ExecCommand(
		whichLxc,
		"profile",
		"device",
		"add",
		vm.Settings.Profile,
		"root",
		"disk",
		"path=/",
		"pool="+vm.Settings.StoragePool.Name)

	vm.Settings.Status = "active"
	err = UpdateBraveSettings(*vm.Settings)
	if err != nil {
		return err
	}
	return nil
}

func checkLXDVersion(whichLxc string) (clientVersion int, serverVersion int, err error) {

	ver, err := shared.ExecCommandWReturn(
		whichLxc,
		"version")

	if err != nil {
		return clientVersion, serverVersion, errors.New("cannot get LXD version")
	}

	v := strings.Split(ver, "\n")
	clientVersionString := strings.TrimSpace(strings.ReplaceAll(strings.Split(v[0], ":")[1], ".", ""))
	serverVersionString := strings.TrimSpace(strings.ReplaceAll(strings.Split(v[0], ":")[1], ".", ""))
	if len(clientVersionString) == 2 {
		clientVersionString = clientVersionString + "0"
	}
	if len(serverVersionString) == 2 {
		serverVersionString = serverVersionString + "0"
	}
	clientVersion, err = strconv.Atoi(clientVersionString)
	if err != nil {
		fmt.Println(err)
	}
	serverVersion, err = strconv.Atoi(serverVersionString)
	if err != nil {
		fmt.Println(err)
	}
	if clientVersion < 303 {
		fmt.Println("Client version: ", clientVersion)
		return clientVersion, serverVersion, errors.New("Bravetools supports LXD >= 3.0.3. Found " + clientVersionString)
	}
	if serverVersion < 303 {
		fmt.Println("Server version: ", serverVersion)
		return serverVersion, serverVersion, errors.New("Bravetools supports LXD >= 3.0.3. Found " + clientVersionString)
	}
	return clientVersion, serverVersion, nil
}

func enableRemote(vm Lxd, whichLxc string) error {
	err := shared.ExecCommand(
		whichLxc,
		"config",
		"set",
		"core.https_address",
		"[::]:8443")
	if err != nil {
		return errors.New("Error connecting to workspace: " + err.Error())
	}

	err = shared.ExecCommand(
		strings.TrimSpace(whichLxc),
		"config",
		"set",
		"core.trust_password",
		vm.Settings.Trust)
	if err != nil {
		return errors.New("Error setting workspace security: " + err.Error())
	}

	return nil
}

// Info ..
func (vm Lxd) Info() (Info, error) {

	backendInfo := Info{}

	_, whichLxc, err := lxdCheck(vm)
	if err != nil {
		return backendInfo, errors.New("Failed to identify LXD: " + err.Error())
	}

	name, err := os.Hostname()
	if err != nil {
		return backendInfo, errors.New("Failed to get host name: " + err.Error())
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return backendInfo, errors.New("Failed to establish UDP connection: " + err.Error())
	}

	defer conn.Close()
	ipv4 := conn.LocalAddr().(*net.UDPAddr).String()
	ipv4 = strings.SplitN(ipv4, ":", -1)[0]

	backendInfo.Name = name
	backendInfo.State = "Running"
	backendInfo.IPv4 = ipv4
	backendInfo.Release = ""
	backendInfo.ImageHash = ""
	backendInfo.Load = ""
	backendInfo.Disk = StorageUsage{}
	backendInfo.Memory = StorageUsage{}

	storageInfo, err := shared.ExecCommandWReturn(whichLxc,
		"storage",
		"info",
		vm.Settings.StoragePool.Name,
		"--bytes")

	if err != nil {
		return backendInfo, errors.New("Unable to access host disk usage: " + err.Error())
	}

	scanner := bufio.NewScanner(strings.NewReader(storageInfo))
	var totalDisk string
	var usedDisk string

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ": ")
		if len(parts) > 1 {
			switch parts[0] {
			case "  space used":
				usedDisk = parts[1]

			case "  total space":
				totalDisk = parts[1]
			}
		}

	}

	usedDisk = usedDisk[1 : len(usedDisk)-1]
	totalDisk = totalDisk[1 : len(totalDisk)-1]
	usedDiskInt, err := strconv.ParseInt(usedDisk, 0, 64)
	if err != nil {
		return backendInfo, err
	}

	totalDiskInt, err := strconv.ParseInt(totalDisk, 0, 64)
	if err != nil {
		return backendInfo, err
	}

	usedDisk = shared.FormatByteCountSI(usedDiskInt)
	totalDisk = shared.FormatByteCountSI(totalDiskInt)

	backendInfo.Disk = StorageUsage{usedDisk, totalDisk}

	totalMemCmd := "cat /proc/meminfo | grep MemTotal | awk '{print $2}'"
	availableMemCmd := "cat /proc/meminfo | grep MemAvailable | awk '{print $2}'"

	totalMem, err := shared.ExecCommandWReturn("sh", "-c", totalMemCmd)
	if err != nil {
		return backendInfo, errors.New("cannot assess total RAM count")
	}
	availableMem, err := shared.ExecCommandWReturn("sh", "-c", availableMemCmd)

	if err != nil {
		return backendInfo, errors.New("cannot assess available RAM count")
	}

	totalMemInt, err := strconv.ParseInt(strings.TrimSpace(totalMem), 0, 64)
	if err != nil {
		return backendInfo, err
	}
	availableMemInt, err := strconv.ParseInt(strings.TrimSpace(availableMem), 0, 64)
	if err != nil {
		return backendInfo, err
	}

	usedMemInt := totalMemInt - availableMemInt

	totalMem = shared.FormatByteCountSI(totalMemInt * 1000)
	usedMem := shared.FormatByteCountSI(usedMemInt * 1000)

	backendInfo.Memory = StorageUsage{usedMem, totalMem}

	cpuCmd := "grep -c ^processor /proc/cpuinfo"
	cpu, err := shared.ExecCommandWReturn("sh",
		"-c",
		cpuCmd)
	if err != nil {
		return backendInfo, errors.New("cannot assess CPU count")
	}

	backendInfo.CPU = cpu

	return backendInfo, nil
}
