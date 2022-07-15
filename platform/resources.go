package platform

import (
	"errors"
	"os"
	"strings"

	"github.com/bravetools/bravetools/shared"
)

// CheckResources ..
func CheckResources(image string, backend Backend, unitParams *shared.Bravefile, bh *BraveHost) error {

	fi, err := os.Stat(image)
	if err != nil {
		return err
	}

	requestedImageSize := fi.Size()

	info, err := backend.Info()
	if err != nil {
		return err
	}

	usedDiskSize, err := shared.SizeCountToInt(info.Disk[0])
	if err != nil {
		return err
	}
	totalDiskSize, err := shared.SizeCountToInt(info.Disk[1])
	if err != nil {
		return err
	}

	if requestedImageSize*5 > (totalDiskSize - usedDiskSize) {
		return errors.New("requested unit size exceeds available disk space on bravetools host. To increase storage pool size modify $HOME/.bravetools/config.yml and run brave configure")
	}

	usedMemorySize, err := shared.SizeCountToInt(info.Memory[0])
	if err != nil {
		return err
	}
	totalMemorySize, err := shared.SizeCountToInt(info.Memory[1])
	if err != nil {
		return err
	} 

	requestedMemorySize, err := shared.SizeCountToInt(unitParams.PlatformService.Resources.RAM)
	if err != nil {
		return err
	}

	if requestedMemorySize > (totalMemorySize - usedMemorySize) {
		return errors.New("Requested unit memory (" + unitParams.PlatformService.Resources.RAM + ") exceeds available memory on bravetools host")
	}

	// Networking Checks
	hostInfo, err := backend.Info()
	if err != nil {
		return errors.New("Failed to connect to host: " + err.Error())
	}

	hostIP := hostInfo.IPv4
	ports := unitParams.PlatformService.Ports
	var hostPorts []string
	if len(ports) > 0 {
		for _, p := range ports {
			ps := strings.Split(p, ":")
			if len(ps) < 2 {
				return errors.New("invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
			}
			hostPorts = append(hostPorts, ps[1])
		}
	}

	err = shared.TCPPortStatus(hostIP, hostPorts)
	if err != nil {
		return err
	}

	return nil
}
