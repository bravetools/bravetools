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
	bravetoolsHome := path.Join(userHome, ".bravetools")
	err := os.RemoveAll(bravetoolsHome)
	return err
}

func checkBravetoolsHome() bool {
	userHome, _ := os.UserHomeDir()
	bravetoolsHome := path.Join(userHome, ".bravetools")
	checkPath, _ := shared.CheckPath(bravetoolsHome)
	return checkPath
}

func lxdCheck(vm Lxd) (LxdStatus, string, error) {
	r, err := shared.ExecCommandWReturn(
		"which",
		"lxc")
	if err != nil {
		return -1, "", err
	}

	if r == "" {
		return NotInstalled, "", nil
	}

	whichLxc := strings.TrimSpace(r)

	if len(whichLxc) > 0 && vm.Settings.Status == "inactive" {
		return NotInitialised, whichLxc, nil
	}

	return Installed, whichLxc, nil
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
		return errors.New("Failed to identify LXD: " + err.Error())
	}

	fmt.Println("LXD status: ", lxdStatus)

	switch lxdStatus {
	case Incompatible:
		_ = deleteBraveHome()
		return errors.New("Incompatible LXD version")
	case NotInstalled:
		_ = deleteBraveHome()
		return errors.New("LXD notinstalled")
	case NotInitialised:
		err = initiateLxd(vm, whichLxc)
		if err != nil {
			_ = deleteBraveHome()
			return errors.New("Failed to initiate Lxd: " + err.Error())
		}

		err = enableRemote(vm, whichLxc)
		if err != nil {
			_ = deleteBraveHome()
			return errors.New("Failed to enable remote: " + err.Error())
		}

		return nil
	case Installed:
		return errors.New("Bravetools is already initialised. Run \"brave configure\" if you'd like to tweak configuration")

	default:
		return nil
	}
}

func initiateLxd(vm Lxd, whichLxc string) error {

	fmt.Println("Createing profile ...")

	err := shared.ExecCommand(
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
		return errors.New("Failed to create storage pool: " + err.Error())
	}

	bridge := "ipv4.address=" + vm.Settings.Network.Bridge + "/24"

	err = shared.ExecCommand(
		whichLxc,
		"network",
		"create",
		"lxdbr0",
		"ipv6.address=none",
		bridge,
		"ipv4.nat=true")
	if err != nil {
		return errors.New("Failed to create network: " + err.Error())
	}

	err = shared.ExecCommand(
		whichLxc,
		"network",
		"attach-profile",
		"lxdbr0",
		"brave",
		"eth0")
	if err != nil {
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
	backendInfo.Disk = []string{}
	backendInfo.Memory = []string{}

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
	totalDiskInt, err := strconv.ParseInt(totalDisk, 0, 64)

	usedDisk = shared.FormatByteCountSI(usedDiskInt)
	totalDisk = shared.FormatByteCountSI(totalDiskInt)

	backendInfo.Disk = []string{usedDisk, totalDisk}

	totalMemCmd := "cat /proc/meminfo | grep MemTotal | awk '{print $2}'"
	availableMemCmd := "cat /proc/meminfo | grep MemAvailable | awk '{print $2}'"

	totalMem, err := shared.ExecCommandWReturn("bash", "-c", totalMemCmd)
	if err != nil {
		return backendInfo, errors.New("Cannot assess total RAM count")
	}
	availableMem, err := shared.ExecCommandWReturn("bash", "-c", availableMemCmd)

	if err != nil {
		return backendInfo, errors.New("Cannot assess available RAM count")
	}

	totalMemInt, err := strconv.ParseInt(strings.TrimSpace(totalMem), 0, 64)
	availableMemInt, err := strconv.ParseInt(strings.TrimSpace(availableMem), 0, 64)
	usedMemInt := totalMemInt - availableMemInt

	totalMem = shared.FormatByteCountSI(totalMemInt * 1000)
	usedMem := shared.FormatByteCountSI(usedMemInt * 1000)

	backendInfo.Memory = []string{usedMem, totalMem}

	cpuCmd := "grep -c ^processor /proc/cpuinfo"
	cpu, err := shared.ExecCommandWReturn("bash",
		"-c",
		cpuCmd)
	if err != nil {
		return backendInfo, errors.New("Cannot assess CPU count")
	}

	backendInfo.CPU = cpu

	return backendInfo, nil
}
