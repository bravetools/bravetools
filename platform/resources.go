package platform

import (
	"errors"
	"strings"

	"github.com/bravetools/bravetools/shared"
)

// CheckResources ..
func CheckResources(image string, backend Backend, unitParams *shared.Service, bh *BraveHost) error {
	info, err := backend.Info()
	if err != nil {
		return errors.New("failed to connect to host: " + err.Error())
	}

	requestedImageSize, err := localImageSize(image)
	if err != nil {
		return err
	}
	freeDiskSpace, err := getFreeSpace(info.Disk)
	if err != nil {
		return err
	}

	if requestedImageSize*5 > freeDiskSpace {
		return errors.New("requested unit size exceeds available disk space on bravetools host. To increase storage pool size modify $HOME/.bravetools/config.yml and run brave configure")
	}

	requestedMemorySize, err := shared.SizeCountToInt(unitParams.Resources.RAM)
	if err != nil {
		return err
	}
	freeMemorySize, err := getFreeSpace(info.Memory)
	if err != nil {
		return err
	}

	if requestedMemorySize > freeMemorySize {
		return errors.New("Requested unit memory (" + unitParams.Resources.RAM + ") exceeds available memory on bravetools host")
	}

	// Networking Checks
	hostIP := info.IPv4
	ports := unitParams.Ports
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

func getFreeSpace(storageUsage StorageUsage) (freeSpace int64, err error) {
	usedStorageBytes, err := shared.SizeCountToInt(storageUsage.UsedStorage)
	if err != nil {
		return freeSpace, errors.New("failed to retrieve backend disk usage:" + err.Error())
	}
	totalStorageBytes, err := shared.SizeCountToInt(storageUsage.TotalStorage)
	if err != nil {
		return freeSpace, errors.New("failed to retrieve backend disk space:" + err.Error())
	}

	return totalStorageBytes - usedStorageBytes, nil
}

// CheckBackendDiskSpace checks whether backend has enough disk space for requested allocation
func CheckBackendDiskSpace(backend Backend, requestedSpace int64) (err error) {
	info, err := backend.Info()
	if err != nil {
		return errors.New("Failed to connect to host: " + err.Error())
	}

	freeSpace, err := getFreeSpace(info.Disk)
	if err != nil {
		return err
	}

	if requestedSpace >= freeSpace {
		return errors.New("requested unit size exceeds available disk space on bravetools host. To increase storage pool size modify $HOME/.bravetools/config.yml and run brave configure")
	}

	return nil
}
