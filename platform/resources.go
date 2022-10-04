package platform

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strings"

	"github.com/bravetools/bravetools/shared"
	lxd "github.com/lxc/lxd/client"
)

// CheckMemory checks if the LXD server host has sufficient RAM to deploy requested unit
func CheckMemory(lxdServer lxd.InstanceServer, ramString string) error {
	resources, err := lxdServer.GetServerResources()
	if err != nil {
		return errors.New("failed to retrieve LXD server resources: " + err.Error())
	}

	requestedMemorySize, err := shared.SizeCountToInt(ramString)
	if err != nil {
		return err
	}
	freeMemorySize := resources.Memory.Total - resources.Memory.Used

	if uint64(requestedMemorySize) > freeMemorySize {
		return errors.New("Requested unit memory (" + ramString + ") exceeds available memory on bravetools host")
	}

	return nil
}

// CheckHostPorts ensures required forwarded ports are free by attempting to connect.
// If a connection is established the port is already taken
func CheckHostPorts(hostURL string, forwardedPorts []string) (err error) {
	parsedURL, err := url.Parse(hostURL)
	if err != nil {
		return fmt.Errorf("failed to parse host URL %q: %s", hostURL, err)
	}
	host, _, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return fmt.Errorf("failed to parse host URL %q: %s", hostURL, err)
	}

	// Networking Checks
	var hostPorts []string
	if len(forwardedPorts) > 0 {
		for _, p := range forwardedPorts {
			ps := strings.Split(p, ":")
			if len(ps) < 2 {
				return errors.New("invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
			}
			hostPorts = append(hostPorts, ps[1])
		}
	}

	err = shared.TCPPortStatus(host, hostPorts)
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

func CheckStoragePoolSpace(lxdServer lxd.InstanceServer, storagePool string, requestedSpace int64) (err error) {
	res, err := lxdServer.GetStoragePoolResources(storagePool)
	if err != nil {
		return fmt.Errorf("failed to retrieve storage information for storage pool %q from lxd server", storagePool)
	}
	freeSpace := res.Space.Total - res.Space.Used

	if freeSpace <= uint64(requestedSpace) {
		// Potential uint64 -> int64 int overflow - just setting to max for now
		if freeSpace > math.MaxInt64 {
			freeSpace = math.MaxInt64
		}
		return fmt.Errorf("requested size exceeds available space on storage pool - %q requested but %q available",
			shared.FormatByteCountSI(requestedSpace), shared.FormatByteCountSI(int64(freeSpace)))
	}

	return nil
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
