package platform

import (
	"bufio"
	"errors"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

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

func lxdCheck(vm Lxd) LxdStatus {
	r, _ := shared.ExecCommandWReturn(
		"which",
		"lxd")

	if r == "" {
		return NotInstalled
	}
	if strings.Compare(r, "/snap/bin/lxd") == 1 {
		if vm.Settings.Status == "inactive" {
			return NotInitialised
		}

		return Installed
	}
	if strings.Compare(r, "/snap/bin/lxd") != 1 {
		return Incompatible
	}

	return -1
}

// NewLxd constructor
func NewLxd(settings HostSettings) *Lxd {
	return &Lxd{
		Settings: &settings,
	}
}

// BraveBackendInit ..
func (vm Lxd) BraveBackendInit() error {
	lxdStatus := lxdCheck(vm)

	switch lxdStatus {
	case Incompatible:
		return errors.New("Incompatible LXD version")
	case NotInstalled:
		err := shared.ExecCommand(
			"sudo",
			"apt",
			"update")
		if err != nil {
			return errors.New("Failed to update workspace: " + err.Error())
		}

		err = installLxd(vm)
		if err != nil {
			return errors.New("Failed to install LXD: " + err.Error())
		}

		err = updateStoragePool(vm)
		if err != nil {
			return errors.New("Failed to update storage pool: " + err.Error())
		}

		err = initiateLxd(vm)
		if err != nil {
			return errors.New("Failed to initiate Lxd: " + err.Error())
		}

		err = enableRemote(vm)
		if err != nil {
			return errors.New("Failed to enable remote: " + err.Error())
		}

		return nil
	case NotInitialised:
		err := updateStoragePool(vm)
		if err != nil {
			return errors.New("Failed to update storage pool: " + err.Error())
		}

		err = initiateLxd(vm)
		if err != nil {
			return errors.New("Failed to initiate Lxd: " + err.Error())
		}

		err = enableRemote(vm)
		if err != nil {
			return errors.New("Failed to enable remote: " + err.Error())
		}

		return nil
	case Installed:
		return errors.New("Bravetools is already initialised. Run \"brave configure\" if you'd like to tweak configuration")

	default:
		return nil
	}
}

func installLxd(vm Lxd) error {
	err := shared.ExecCommand(
		"sudo",
		"snap",
		"install",
		"--stable",
		"lxd")
	if err != nil {
		return err
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}
	err = shared.ExecCommand(
		"sudo",
		"usermod",
		"-aG",
		"lxd",
		usr.Username)
	if err != nil {
		return err
	}
	return nil
}

func updateStoragePool(vm Lxd) error {
	timestamp := time.Now()
	storagePoolName := vm.Settings.StoragePool.Name + "-" + timestamp.Format("20060102150405")
	vm.Settings.StoragePool.Name = storagePoolName

	err := UpdateBraveSettings(*vm.Settings)
	if err != nil {
		return err
	}

	return nil
}

func initiateLxd(vm Lxd) error {
	var lxdInit = `cat <<EOF | sudo lxd init --preseed
pools:
- name: ` + vm.Settings.StoragePool.Name + "\n" +
		`  driver: zfs
networks:
- name: lxdbr0
  type: bridge
  config:` + "\n" +
		"    ipv4.address: " + vm.Settings.Network.Bridge + "/24 \n" +
		`    ipv4.nat: true
    ipv6.address: none
profiles:
- name: ` + vm.Settings.Profile + "\n" +
		`  devices:
    root:
      path: /
      pool: ` + vm.Settings.StoragePool.Name + "\n" +
		`      type: disk
    eth0:
      nictype: bridged
      parent: lxdbr0
      type: nic
EOF`
	err := shared.ExecCommand(
		shared.SnapLXC,
		"profile",
		"create",
		vm.Settings.Profile)
	if err != nil {
		return errors.New("Failed to create LXD profile: " + err.Error())
	}

	err = shared.ExecCommand(
		shared.SnapLXC,
		"storage",
		"create",
		vm.Settings.StoragePool.Name,
		vm.Settings.StoragePool.Type,
		"size="+vm.Settings.StoragePool.Size)
	if err != nil {
		return errors.New("Failed to create storage pool: " + err.Error())
	}

	shared.ExecCommand(
		shared.SnapLXC,
		"profile",
		"device",
		"add",
		vm.Settings.Profile,
		"root",
		"disk",
		"path=/",
		"pool="+vm.Settings.StoragePool.Name)

	err = shared.ExecCommand(
		"bash",
		"-c",
		lxdInit)
	if err != nil {
		return errors.New("Failed to initiate workspace: " + err.Error())
	}

	vm.Settings.Status = "active"
	err = UpdateBraveSettings(*vm.Settings)
	if err != nil {
		return err
	}
	return nil
}

func enableRemote(vm Lxd) error {
	err := shared.ExecCommand(
		shared.SnapLXC,
		"config",
		"set",
		"core.https_address",
		"[::]:8443")
	if err != nil {
		return errors.New("Error connecting to workspace: " + err.Error())
	}

	err = shared.ExecCommand(
		shared.SnapLXC,
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

	storageInfo, err := shared.ExecCommandWReturn("bash", "-c",
		shared.SnapLXC+" storage info "+vm.Settings.StoragePool.Name+" --bytes")

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
