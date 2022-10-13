package platform

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bravetools/bravetools/shared"
	"github.com/briandowns/spinner"
	"github.com/mitchellh/go-ps"
)

type (
	// Multipass type defines local dev VM
	Multipass struct {
		Settings HostSettings
	}
)

// NewMultipass constructor
func NewMultipass(settings HostSettings) *Multipass {
	return &Multipass{
		Settings: settings,
	}
}

// checkMultipass checks if Multipass is running
func checkMultipass() (bool, error) {

	ps, err := ps.Processes()
	if err != nil {
		return false, err
	}

	found := false
	for _, p := range ps {
		if strings.Contains(p.Executable(), "multipass") {
			found = true
			break
		}
	}

	if !found {
		return false, errors.New("install multipass")
	}

	return true, nil
}

// BraveBackendInit creates a new instance of BraveAI host
func (vm Multipass) BraveBackendInit() error {

	_, err := checkMultipass()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		err = shared.ExecCommand("multipass",
			"set",
			"local.privileged-mounts=Yes")
		if err != nil {
			log.Println("failed to set local.privileged-mounts to true, attempting to continue")
		}
	}

	err = shared.ExecCommand("multipass",
		"launch",
		"--cpus",
		vm.Settings.BackendSettings.Resources.CPU,
		"--disk",
		vm.Settings.BackendSettings.Resources.HD,
		"--mem",
		vm.Settings.BackendSettings.Resources.RAM,
		"--name",
		vm.Settings.BackendSettings.Resources.Name,
		vm.Settings.BackendSettings.Resources.OS)
	if err != nil {
		return errors.New("failed to create workspace: " + err.Error())
	}

	time.Sleep(10 * time.Second)

	usr, err := user.Current()
	if err != nil {
		return errors.New("unable to fetch current user information: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"snap",
		"install",
		"multipass-sshfs")

	if err != nil {
		return errors.New("failed to update workspace: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"mount",
		filepath.Join(usr.HomeDir, shared.BraveHome),
		vm.Settings.Name+":/home/ubuntu"+shared.BraveHome)

	if err != nil {
		return errors.New("unable to mount local volumes to multipass: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"apt",
		"update")

	if err != nil {
		return errors.New("failed to update workspace: " + err.Error())
	}

	shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"apt",
		"remove",
		"-y",
		"lxd")
	shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"apt",
		"autoremove",
		"-y")
	shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"apt",
		"purge")

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"snap",
		"install",
		"--stable",
		"lxd")
	if err != nil {
		return errors.New("unable to install LXD: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"sudo",
		"usermod",
		"-aG",
		"lxd",
		"ubuntu")

	if err != nil {
		return errors.New("failed to install packages in workspace: " + err.Error())
	}

	fmt.Println("Installing required software ...")
	time.Sleep(10 * time.Second)

	timestamp := time.Now()
	storagePoolName := vm.Settings.Profile + "-" + timestamp.Format("20060102150405")
	vm.Settings.StoragePool.Name = storagePoolName

	err = UpdateBraveSettings(vm.Settings)
	if err != nil {
		return errors.New("failed update settings" + err.Error())
	}

	var lxdInit = `cat <<EOF | sudo lxd init --preseed
pools:
- name: ` + vm.Settings.StoragePool.Name + "\n" +
		`  driver: zfs
networks:
- name: ` + vm.Settings.Network.Name + "\n" +
		`  type: bridge
  config:` + "\n" +
		"    ipv4.address: " + vm.Settings.Network.IP + "/24 \n" +
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
      parent: ` + vm.Settings.Network.Name + "\n" +
		`      type: nic
EOF`

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		shared.SnapLXC,
		"profile",
		"create",
		vm.Settings.Profile)
	if err != nil {
		return errors.New("failed to create LXD profile: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		shared.SnapLXC,
		"storage",
		"create",
		vm.Settings.StoragePool.Name,
		vm.Settings.StoragePool.Type,
		"size="+vm.Settings.StoragePool.Size)
	if err != nil {
		return errors.New("failed to create storage pool: " + err.Error())
	}

	shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		shared.SnapLXC,
		"profile",
		"device",
		"add",
		vm.Settings.Profile,
		"root",
		"disk",
		"path=/",
		"pool="+vm.Settings.StoragePool.Name)

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"bash",
		"-c",
		lxdInit)
	if err != nil {
		return errors.New("failed to initiate workspace: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		shared.SnapLXC,
		"config",
		"set",
		"core.https_address",
		"[::]:8443")
	if err != nil {
		return errors.New("error connecting to workspace: " + err.Error())
	}

	err = shared.ExecCommand("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		shared.SnapLXC,
		"config",
		"set",
		"core.trust_password",
		vm.Settings.Trust)
	if err != nil {
		return errors.New("error setting workspace security: " + err.Error())
	}

	vm.Settings.Status = "active"
	err = UpdateBraveSettings(vm.Settings)
	if err != nil {
		return err
	}
	return nil

}

// BraveHostDelete removes BraveAI host
func (vm Multipass) BraveHostDelete() error {

	err := shared.ExecCommand("multipass", "delete", vm.Settings.Name)
	if err != nil {
		return err
	}
	err = shared.ExecCommand("multipass", "purge")
	if err != nil {
		return err
	}

	return nil
}

// Info shows all VMs and their state
func (vm Multipass) Info() (backendInfo Info, err error) {
	operation := shared.Info("Gathering multipass settings")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()
	defer s.Stop()

	_, err = checkMultipass()

	if err != nil {
		return backendInfo, errors.New("multipass process not found: " + err.Error())
	}

	backendInfo, err = vm.getInfo()
	if err != nil {
		return backendInfo, errors.New("error contacting multipass vm: " + err.Error())
	}

	if backendInfo.State == "Running" {

		doneChan := make(chan struct{}, 3)
		errChan := make(chan error)

		go func() {
			backendInfo.Disk, err = vm.getDiskUsage()
			if err != nil {
				errChan <- errors.New("unable to access host disk usage: " + err.Error())
			} else {
				doneChan <- struct{}{}
			}
		}()
		go func() {
			backendInfo.Memory, err = vm.getRamUsage()
			if err != nil {
				errChan <- errors.New("cannot assess total RAM count: " + err.Error())
			} else {
				doneChan <- struct{}{}
			}
		}()
		go func() {
			backendInfo.CPU, err = vm.getCpuCount()
			if err != nil {
				errChan <- errors.New("cannot assess CPU count: " + err.Error())
			} else {
				doneChan <- struct{}{}
			}
		}()

		for i := 0; i < cap(doneChan); i++ {
			select {
			case <-errChan:
				return backendInfo, err
			case <-doneChan:
			}
		}

	}
	return backendInfo, nil
}

// Info shows all VMs and their state
func (vm Multipass) Running() (bool, error) {
	operation := shared.Info("Gathering multipass settings")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()
	defer s.Stop()

	_, err := checkMultipass()

	if err != nil {
		return false, errors.New("multipass process not found: " + err.Error())
	}

	backendInfo, err := vm.getInfo()
	if err != nil {
		return false, errors.New("error contacting multipass vm: " + err.Error())
	}

	if backendInfo.State == "Running" {
		return true, nil
	}
	return false, nil
}

// Start starts the backend if it is not already running
func (vm Multipass) Start() error {
	// Actually starting the VM takes longer compared to confirming that it is still running
	// This longer running process should be messaged to the user else it looks like the program hangs
	// However, short pauses when VM is actually running make the message flash and dissapear, creating noise.

	// Below we ensure the VM is started - only if the process takes longer than 1 second is the spinner shown

	done := make(chan struct{})
	var err error

	go func() {
		defer close(done)
		cmd := exec.Command("multipass", "start", vm.Settings.Name)
		err = cmd.Run()
	}()

	// Race completion of above command against 1 second timer
	// If 1 second passes before completion of command, start spinner and then await completion of command
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		operation := shared.Info("Ensuring multipass VM is started")
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		s.Suffix = " " + operation
		s.Start()
		defer s.Stop()
		<-done
	}

	if err != nil {
		return err
	}

	return nil
}

func (vm Multipass) getInfo() (Info, error) {

	backendInfo := NewInfo()

	out, err := exec.Command("multipass", "info", vm.Settings.Name).Output()
	if err != nil {
		return backendInfo, err
	}

	info := strings.Split(string(out), "\n")
	for _, data := range info {
		d := strings.Split(data, ":")
		key := strings.TrimSpace(d[0])
		switch key {
		case "Name":
			backendInfo.Name = strings.TrimSpace(d[1])
		case "State":
			backendInfo.State = strings.TrimSpace(d[1])
		case "IPv4":
			backendInfo.IPv4 = strings.TrimSpace(d[1])
		case "Release":
			backendInfo.Release = strings.TrimSpace(d[1])
		case "Image hash":
			backendInfo.ImageHash = strings.TrimSpace(d[1])
		case "Load":
			backendInfo.Load = strings.TrimSpace(d[1])
		}
	}

	return backendInfo, nil
}

func (vm Multipass) getDiskUsage() (storage StorageUsage, err error) {
	cmd := shared.SnapLXC + " storage info " + vm.Settings.StoragePool.Name + " --bytes"
	storageInfo, err := shared.ExecCommandWReturn("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"bash", "-c",
		cmd)

	if err != nil {
		return storage, err
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
		return storage, err
	}

	totalDiskInt, err := strconv.ParseInt(totalDisk, 0, 64)
	if err != nil {
		return storage, err
	}

	usedDisk = shared.FormatByteCountSI(usedDiskInt)
	totalDisk = shared.FormatByteCountSI(totalDiskInt)

	storage = StorageUsage{usedDisk, totalDisk}
	return storage, nil
}

func (vm Multipass) getRamUsage() (storage StorageUsage, err error) {
	totalMemCmd := "cat /proc/meminfo | grep MemTotal | awk '{print $2}'"
	availableMemCmd := "cat /proc/meminfo | grep MemAvailable | awk '{print $2}'"

	totalMem, err := shared.ExecCommandWReturn("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"bash", "-c", totalMemCmd)
	if err != nil {
		return storage, err
	}

	totalMem = strings.Split(strings.TrimSpace(strings.Split(totalMem, ":")[1]), " ")[0]

	availableMem, err := shared.ExecCommandWReturn("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"bash", "-c", availableMemCmd)

	if err != nil {
		return storage, err
	}

	availableMem = strings.Split(strings.TrimSpace(strings.Split(availableMem, ":")[1]), " ")[0]

	totalMemInt, err := strconv.Atoi(totalMem)
	if err != nil {
		return storage, err
	}

	availableMemInt, err := strconv.Atoi(availableMem)
	if err != nil {
		return storage, err
	}

	usedMemInt := totalMemInt - availableMemInt

	totalMem = shared.FormatByteCountSI(int64(totalMemInt * 1000))
	usedMem := shared.FormatByteCountSI(int64(usedMemInt * 1000))

	storage = StorageUsage{usedMem, totalMem}
	return storage, nil
}

func (vm Multipass) getCpuCount() (cpu string, err error) {
	cpuCount := "grep -c ^processor /proc/cpuinfo"
	cpu, err = shared.ExecCommandWReturn("multipass",
		"exec",
		vm.Settings.Name,
		"--",
		"bash",
		"-c",
		cpuCount)

	return cpu, err
}
