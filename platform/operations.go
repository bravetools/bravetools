package platform

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bravetools/bravetools/shared"

	pem "encoding/pem"

	"github.com/briandowns/spinner"

	lxd "github.com/canonical/lxd/client"
	lxdshared "github.com/canonical/lxd/shared"
	api "github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/ioprogress"
)

// DeleteNetwork ..
func DeleteNetwork(lxdServer lxd.InstanceServer, name string) error {
	err := lxdServer.DeleteNetwork(name)
	if err != nil {
		return errors.New("Failed to delete Brave profile: " + err.Error())
	}

	return nil
}

// DeleteProfile ..
func DeleteProfile(lxdServer lxd.InstanceServer, name string) error {
	err := lxdServer.DeleteProfile(name)
	if err != nil {
		return errors.New("Failed to delete Brave profile: " + err.Error())
	}

	return nil
}

// DeleteStoragePool ..
func DeleteStoragePool(lxdServer lxd.InstanceServer, name string) error {
	err := lxdServer.DeleteStoragePool(name)
	if err != nil {
		return errors.New("Failed to delete Brave storage pool: " + err.Error())
	}

	return nil
}

// SetActiveStoragePool pool assigns a profile with default storage
func SetActiveStoragePool(lxdServer lxd.InstanceServer, name string) error {
	profileName := shared.BravetoolsVmName

	profile, etag, err := lxdServer.GetProfile(profileName)
	if err != nil {
		return errors.New("Unable to load profile: " + err.Error())
	}

	device := map[string]string{}

	device["type"] = "disk"
	device["path"] = "/"
	device["pool"] = name

	profile.Devices["root"] = device

	err = lxdServer.UpdateProfile(profileName, profile.Writable(), etag)
	if err != nil {
		return errors.New("Failed to update profile with storage pool configuration: " + err.Error())
	}

	return nil
}

// CreateStoragePool creates a new storage pool
func CreateStoragePool(lxdServer lxd.InstanceServer, name string, size string) error {
	req := api.StoragePoolsPost{
		Name:   name,
		Driver: "zfs",
	}

	req.Config = map[string]string{
		"size": size,
	}

	err := lxdServer.CreateStoragePool(req)
	if err != nil {
		return errors.New("failed to create storage pool: " + err.Error())
	}

	return nil
}

// AddRemote adds remote LXC host
func AddRemote(remote Remote, password string) error {
	var err error
	userHome, _ := os.UserHomeDir()
	certf := path.Join(userHome, shared.BraveClientCert)
	keyf := path.Join(userHome, shared.BraveClientKey)

	options := lxdshared.CertOptions{}
	options.AddHosts = false
	options.CommonName = ""
	options.SubjectAlternativeNames = []string{}

	// Generate client certificates
	err = lxdshared.FindOrGenCert(certf, keyf, true, options)
	if err != nil {
		return err
	}

	if remote.Protocol == "unix" {
		return nil
	}

	// Check if the system CA worked for the TLS connection
	var certificate *x509.Certificate
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	certificate, err = lxdshared.GetRemoteCertificate(ctx, remote.URL, "")
	if err != nil {
		return err
	}
	if certificate == nil {
		return errors.New("failed to get lxd certificate - certificate is nil")
	}

	// LXD may be running in a VM and check certificate validity based on VM clock causing issues
	// Waiting a few seconds gives a safety buffer to prevent sync issues
	waitTime, err := time.ParseDuration("5s")
	if err != nil {
		return err
	}
	time.Sleep(waitTime)

	// Handle certificate prompt
	digest := lxdshared.CertFingerprint(certificate)
	fmt.Printf(("Certificate fingerprint: %s")+"\n", digest)

	dnam := path.Join(userHome, shared.BraveServerCertStore)
	err = os.MkdirAll(dnam, 0750)
	if err != nil {
		return errors.New("could not create server cert dir")
	}

	certf = path.Join(dnam, remote.Name+".crt")
	certOut, err := os.Create(certf)
	if err != nil {
		return err
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	certOut.Close()

	// Load newly generated certs from disk into Remote struct
	remote, err = LoadRemoteSettings(remote.Name)
	if err != nil {
		return fmt.Errorf("failed to load remote %q from disk: %s", remote.Name, err)
	}

	req := api.CertificatesPost{
		Password: password,
	}
	req.Type = "client"

	lxdServer, err := GetLXDInstanceServer(remote)
	if err != nil {
		return err
	}

	// Ensure we are not already trusted by server before adding cert
	server, _, err := lxdServer.GetServer()
	if err != nil {
		return err
	}
	if server.Auth != "trusted" {
		err = lxdServer.CreateCertificate(req)
		if err != nil {
			return err
		}

		// Reconnect and check if now trusted
		lxdServer, err = GetLXDInstanceServer(remote)
		if err != nil {
			return err
		}
		server, _, err = lxdServer.GetServer()
		if err != nil {
			return err
		}
		if server.Auth != "trusted" {
			return errors.New("failed to authenticate with server - still not trusted after adding cert")
		}
	}

	return nil
}

// RemoveRemote removes remote LXC host
func RemoveRemote(name string) error {

	if name == shared.BravetoolsRemote {
		return fmt.Errorf("default bravetools remote %q cannot be removed", name)
	}

	remoteNames, err := ListRemotes()
	if err != nil {
		return err
	}

	if !shared.StringInSlice(name, remoteNames) {
		return errors.New("remote " + name + " does not exist")
	}

	userHome, _ := os.UserHomeDir()
	remotef := path.Join(userHome, shared.BraveRemoteStore, name+".json")
	certs := path.Join(userHome, shared.BraveServerCertStore, name+".crt")

	err = os.Remove(remotef)
	if err != nil {
		return err
	}

	// Remove associated cert if it exists
	if shared.FileExists(certs) {
		err = os.Remove(certs)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteDevice unmounts a disk
func DeleteDevice(lxdServer lxd.InstanceServer, name string, target string) (string, error) {

	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return "", err
	}

	devname := target
	device, ok := inst.Devices[devname]
	if !ok {
		return "", fmt.Errorf("device %q doesn't exist", devname)
	}

	source := device["source"]

	delete(inst.Devices, target)

	op, err := lxdServer.UpdateInstance(name, inst.Writable(), etag)
	if err != nil {
		return "", err
	}

	err = op.Wait()
	if err != nil {
		return "", err
	}

	return source, nil
}

// AddDevice adds an external device to a unit with the given devSettings
func AddDevice(lxdServer lxd.InstanceServer, unitName string, devname string, devSettings map[string]string) error {
	inst, etag, err := lxdServer.GetInstance(unitName)
	if err != nil {
		return errors.New("Error accessing unit: " + unitName)
	}

	inst.Devices[devname] = devSettings

	op, err := lxdServer.UpdateInstance(unitName, inst.Writable(), etag)
	if err != nil {
		return errors.New("Errors updating unit configuration: " + unitName)
	}

	err = op.Wait()
	if err != nil {
		return errors.New("Error updating unit " + unitName + " Error: " + err.Error())
	}

	return nil

}

// UpdateDevice updates the deviceSettings of an existing device - existing config remains unchanged unless
// overwritten by a matching key in the provided deviceSettings
func UpdateDevice(lxdServer lxd.InstanceServer, unitName string, deviceName string, deviceSettings map[string]string) error {
	inst, etag, err := lxdServer.GetInstance(unitName)
	if err != nil {
		return errors.New("Error accessing unit: " + unitName)
	}

	currentSettings, ok := inst.Devices[deviceName]
	if !ok {
		return fmt.Errorf("device %q not found on unit %q", deviceName, unitName)
	}

	for k, v := range deviceSettings {
		currentSettings[k] = v
	}

	op, err := lxdServer.UpdateInstance(unitName, inst.Writable(), etag)
	if err != nil {
		return errors.New("Errors updating unit configuration: " + unitName)
	}

	err = op.Wait()
	if err != nil {
		return errors.New("Error updating unit " + unitName + " Error: " + err.Error())
	}

	return nil
}

// MountDirectory mounts local directory to unit
func MountDirectory(lxdServer lxd.InstanceServer, sourcePath string, destUnit string, destPath string) error {
	inst, etag, err := lxdServer.GetInstance(destUnit)
	if err != nil {
		return err
	}

	sourcePath = path.Clean(sourcePath)

	destPath = cleanMountTargetPath(destPath)
	hashStr := getDiskDeviceHash(destUnit, destPath)

	devname := hashStr
	_, exists := inst.Devices[devname]
	if exists {
		return fmt.Errorf("something is already mounted on target path %q in unit %q", destPath, destUnit)
	}

	device := map[string]string{}
	device["type"] = "disk"
	device["source"] = sourcePath
	device["path"] = destPath

	inst.Devices[devname] = device

	op, err := lxdServer.UpdateInstance(destUnit, inst.Writable(), etag)
	if err != nil {
		return errors.New("failed to update unit settings: " + err.Error())
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// GetImages returns all images from host
func GetImages(lxdServer lxd.ImageServer) ([]api.Image, error) {
	images, err := lxdServer.GetImages()
	if err != nil {
		return nil, err
	}

	return images, nil
}

// DeleteVolume ..
func DeleteVolume(lxdServer lxd.InstanceServer, pool string, volume api.StorageVolume) error {
	err := lxdServer.DeleteStoragePoolVolume(pool, volume.Type, volume.Name)
	if err != nil {
		return errors.New("failed to delete volume: " + err.Error())
	}

	return nil
}

// GetVolume ..
func GetVolume(lxdServer lxd.InstanceServer, pool string) (volume api.StorageVolume, err error) {
	volumes, err := lxdServer.GetStoragePoolVolumes(pool)
	if err != nil {
		return volume, err
	}

	if len(volumes) > 0 {
		for _, v := range volumes {
			if v.Type == "custom" {
				volume = v
				break
			}
		}
	}
	return volume, nil
}

// GetBraveProfile ..
func GetBraveProfile(lxdServer lxd.InstanceServer, profileName string) (braveProfile shared.BraveProfile, err error) {
	srv, _, err := lxdServer.GetServer()
	if err != nil {
		log.Fatal("LXD server error: " + err.Error())
	}
	braveProfile.LxdVersion = srv.Environment.ServerVersion
	pNames, _ := lxdServer.GetProfileNames()

	for _, pName := range pNames {
		if pName == profileName {
			braveProfile.Name = pName
			profile, _, _ := lxdServer.GetProfile(pName)
			for k, v := range profile.Devices {
				if k == "eth0" {
					braveProfile.Bridge = v["parent"]
				}
				if k == "root" {
					braveProfile.Storage = v["pool"]
				}
			}
			return braveProfile, nil
		}
	}
	return braveProfile, errors.New("profile not found")
}

func containerHasProfile(container *api.Container, profileName string) bool {
	for _, p := range container.Profiles {
		if p == profileName {
			return true
		}
	}
	return false
}

// GetUnits returns all running units
func GetUnits(lxdServer lxd.InstanceServer, profileName string) (units []shared.BraveUnit, err error) {
	names, err := lxdServer.GetContainerNames()
	if err != nil {
		return nil, err
	}
	for _, n := range names {
		containerState, _, _ := lxdServer.GetContainerState(n)
		var unit shared.BraveUnit
		container, _, _ := lxdServer.GetContainer(n)

		// Check if selected user profile manages this container
		if !containerHasProfile(container, profileName) {
			continue
		}

		devices := container.Devices
		var diskDevice []shared.DiskDevice
		var disk shared.DiskDevice

		var proxyDevice []shared.ProxyDevice
		var proxy shared.ProxyDevice
		var nicDevice shared.NicDevice

		for k, device := range devices {
			if val, ok := device["type"]; ok {
				switch val {
				case "disk":
					disk.Name = k
					disk.Path = device["path"]
					disk.Source = device["source"]
					diskDevice = append(diskDevice, disk)
				case "proxy":
					proxy.Name = k
					proxy.ConnectIP = device["connect"]
					proxy.ListenIP = device["listen"]
					proxyDevice = append(proxyDevice, proxy)

				case "nic":
					nicDevice.Name = k
					nicDevice.Parent = device["parent"]
					nicDevice.Type = device["type"]
					nicDevice.NicType = device["nictype"]
					nicDevice.IP = device["ipv4.address"]
				}
			}
		}

		unit.Name = n
		unit.Status = containerState.Status
		if strings.ToLower(containerState.Status) == "running" {
			if eth, ok := containerState.Network["eth0"]; ok {
				unit.Address = eth.Addresses[0].Address
			}
		}
		unit.Disk = diskDevice
		unit.Proxy = proxyDevice
		unit.NIC = nicDevice
		units = append(units, unit)
	}

	return units, nil
}

// LaunchFromImage creates new unit based on image
func LaunchFromImage(destServer lxd.InstanceServer, sourceServer lxd.ImageServer, imageName string, containerName string, profileName string, storagePool string) (fingerprint string, err error) {
	operation := shared.Info("Launching " + containerName)
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()
	defer s.Stop()

	destServerArch, err := GetLXDServerArch(destServer)
	if err != nil {
		return fingerprint, err
	}

	req := api.ContainersPost{
		Name: containerName,
	}
	req.Profiles = []string{profileName}

	// Attach a specific disk when launching if requested
	if storagePool != "" {
		if req.Devices == nil {
			req.Devices = make(map[string]map[string]string)
		}
		req.Devices["root"] = map[string]string{
			"path": "/",
			"pool": storagePool,
			"type": "disk",
		}
	}

	fingerprint, err = GetFingerprintByAlias(sourceServer, imageName, destServerArch)
	if err != nil {
		return fingerprint, err
	}

	imgInfo, _, err := sourceServer.GetImage(fingerprint)
	if err != nil {
		return fingerprint, err
	}

	op, err := destServer.CreateContainerFromImage(sourceServer, *imgInfo, req)
	if err != nil {
		return fingerprint, err
	}

	err = op.Wait()
	if err != nil {
		return fingerprint, err
	}

	return fingerprint, nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
		log.SetOutput(os.Stdout)
		log.Println("retrying:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func isIPv4(ip string) bool {
	parts := strings.Split(ip, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

type ExecArgs struct {
	env    map[string]string
	detach bool
}

// Exec runs command inside unit
func Exec(ctx context.Context, lxdServer lxd.InstanceServer, name string, command []string, arg ExecArgs) (returnCode int, err error) {
	if err = ctx.Err(); err != nil {
		return 0, err
	}

	err = retry(10, 4*time.Second, func() (err error) {
		if err = ctx.Err(); err != nil {
			return err
		}
		c, _, err := lxdServer.GetContainerState(name)
		if err != nil {
			return fmt.Errorf("failed to get container %q: %s", name, err.Error())
		}

		fmt.Println("network:", c.Network)

		ip := c.Network["eth0"].Addresses[0].Address
		isIP := isIPv4(ip)
		if !isIP {
			return errors.New("getting IPv6 info")
		}
		return
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return 100, err
	}

	fmt.Println(shared.Info("["+name+"] "+"RUN: "), shared.Warn(command))

	req := api.ContainerExecPost{
		Command:   command,
		WaitForWS: true,
		// record-output variable when set to true records stdout and stderr to disk.
		// wait-for-websocket must be set to false.
		// Both cannot be rue at the same time
		// ! THIS IS A TEMPORARY FIX
		// TODO: record-output and wait-for-websocket should be handled explicitly in Bravefile
		RecordOutput: false,
		Interactive:  false,
		Environment:  arg.env,
	}

	args := lxd.ContainerExecArgs{
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Control:  nil, // terminal non-interactive
		DataDone: make(chan bool),
	}

	// TODO: this parameter should be handled in Bravefile
	if arg.detach {
		req.WaitForWS = false
	}

	op, err := lxdServer.ExecContainer(name, req, &args)

	if err != nil {
		return 1, errors.New("error getting current state: " + err.Error())
	}

	if arg.detach {
		return returnCode, nil
	}

	opWait := make(chan struct{})
	go func() {
		err = op.Wait()
		close(opWait)
	}()

	select {
	case <-ctx.Done():
		return 1, ctx.Err()
	case <-opWait:
	}

	if err != nil {
		return 1, errors.New("error executing command: " + err.Error())
	}
	opAPI := op.Get()

	returnCode = int(opAPI.Metadata["return"].(float64))

	return returnCode, nil
}

// Delete deletes a unit on a LXD remote
func DeleteUnit(lxdServer lxd.InstanceServer, name string) error {
	unit, _, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}

	if unit.Status == "Running" {

		req := api.InstanceStatePut{
			Action:  "stop",
			Timeout: -1,
			Force:   true,
		}

		op, err := lxdServer.UpdateInstanceState(name, req, "")
		if err != nil {
			return err
		}

		err = op.Wait()
		if err != nil {
			return errors.New("stopping the instance failed: " + err.Error())
		}
	}

	op, err := lxdServer.DeleteContainer(name)
	if err != nil {
		return errors.New("fail to delete unit: " + err.Error())
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Start unit
func Start(lxdServer lxd.InstanceServer, name string) error {

	unit, _, err := lxdServer.GetContainer(name)
	if err != nil {
		return err
	}

	state := false

	if unit.Status == "Stopped" {
		req := api.InstanceStatePut{
			Action:   "start",
			Timeout:  -1,
			Force:    true,
			Stateful: state,
		}

		if unit.Stateful {
			state = true
		}

		op, err := lxdServer.UpdateInstanceState(name, req, "")
		if err != nil {
			return err
		}

		err = op.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop unit
func Stop(lxdServer lxd.InstanceServer, name string) error {
	unit, _, err := lxdServer.GetContainer(name)
	if err != nil {
		return err
	}

	if unit.Status == "Running" {
		req := api.InstanceStatePut{
			Action:  "stop",
			Timeout: -1,
			Force:   true,
		}

		op, err := lxdServer.UpdateInstanceState(name, req, "")
		if err != nil {
			return err
		}

		err = op.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

// Publish unit
// lxc publish -f [remote]:[name] [remote]: --alias [image]
func Publish(lxdServer lxd.InstanceServer, name string, image string) (fingerprint string, err error) {
	operation := shared.Info("Publishing " + name)
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()

	unit, _, err := lxdServer.GetInstance(name)
	if err != nil {
		return "", err
	}

	var unitStatus = unit.Status

	if unit.Status == "Running" {
		req := api.InstanceStatePut{
			Action:  "stop",
			Timeout: -1,
			Force:   true,
		}

		op, err := lxdServer.UpdateInstanceState(name, req, "")
		if err != nil {
			return "", err
		}

		err = op.Wait()
		if err != nil {
			return "", err
		}
	}

	// Create image
	req := api.ImagesPost{
		Source: &api.ImagesPostSource{
			Type: "container",
			Name: name,
		},
	}

	op, err := lxdServer.CreateImage(req, nil)
	if err != nil {
		return "", err
	}

	err = op.Wait()
	if err != nil {
		return "", err
	}

	opAPI := op.Get()
	fingerprint = opAPI.Metadata["fingerprint"].(string)

	aliasPost := api.ImageAliasesPost{}
	aliasPost.Name = image
	if aliasPost.Name != "" {
		aliasPost.Name = name
	}
	aliasPost.Target = fingerprint
	err = lxdServer.CreateImageAlias(aliasPost)
	if err != nil {
		return fingerprint, err
	}

	if unitStatus == "Running" {
		req := api.InstanceStatePut{
			Action:  "start",
			Timeout: -1,
			Force:   true,
		}

		op, err := lxdServer.UpdateInstanceState(name, req, "")
		if err != nil {
			return fingerprint, err
		}

		err = op.Wait()
		if err != nil {
			return fingerprint, err
		}

	}

	s.Stop()
	return fingerprint, nil
}

// SymlinkPush  copies a symlink into unit
func SymlinkPush(lxdServer lxd.InstanceServer, name string, sourceFile string, targetPath string) error {
	var readCloser io.ReadCloser

	fi, err := os.Lstat(sourceFile)
	if err != nil {
		return errors.New("Unable to read symlink " + sourceFile + ": " + err.Error())
	}

	symlinkTarget, err := os.Readlink(sourceFile)
	if err != nil {
		return errors.New("Unable to read symlink " + sourceFile + ": " + err.Error())
	}

	mode, uid, gid := lxdshared.GetOwnerMode(fi)
	args := lxd.InstanceFileArgs{
		UID:  int64(uid),
		GID:  int64(gid),
		Mode: int(mode.Perm()),
	}

	args.Type = "symlink"
	args.Content = bytes.NewReader([]byte(symlinkTarget))
	readCloser = ioutil.NopCloser(args.Content)

	fmt.Printf(shared.Info("Pushing %s to %s (%s)\n"), sourceFile, targetPath, args.Type)

	contentLength, err := args.Content.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = args.Content.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	args.Content = lxdshared.NewReadSeeker(&ioprogress.ProgressReader{
		ReadCloser: readCloser,
		Tracker: &ioprogress.ProgressTracker{
			Length: contentLength,
		},
	}, args.Content)

	_, targetFile := filepath.Split(sourceFile)
	err = lxdServer.CreateInstanceFile(name, filepath.Join(targetPath, targetFile), args)
	if err != nil {
		return err
	}

	return nil
}

// FilePush copies local file into unit
func FilePush(lxdServer lxd.InstanceServer, name string, sourceFile string, targetPath string) error {
	var readCloser io.ReadCloser
	fInfo, err := os.Stat(sourceFile)

	if err != nil {
		return errors.New("unable to read file " + sourceFile + ": " + err.Error())
	}

	mode, uid, gid := lxdshared.GetOwnerMode(fInfo)
	args := lxd.InstanceFileArgs{
		UID:  int64(uid),
		GID:  int64(gid),
		Mode: int(mode.Perm()),
	}

	f, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer f.Close()

	args.Type = "file"
	args.Content = f
	readCloser = f

	contentLength, err := args.Content.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = args.Content.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	args.Content = lxdshared.NewReadSeeker(&ioprogress.ProgressReader{
		ReadCloser: readCloser,
		Tracker: &ioprogress.ProgressTracker{
			Length: contentLength,
		},
	}, args.Content)

	fmt.Printf(shared.Info("| Pushing %s to %s (%s)\n"), sourceFile, targetPath, args.Type)

	_, targetFile := filepath.Split(sourceFile)

	target := filepath.Join(targetPath, targetFile)

	hostOs := runtime.GOOS
	if hostOs == "windows" {
		target = strings.Replace(target, string(filepath.Separator), "/", -1)
	}

	err = lxdServer.CreateInstanceFile(name, target, args)
	if err != nil {
		return err
	}

	return nil
}

// ImportImage imports image from current directory
func ImportImage(lxdServer lxd.InstanceServer, imageTar string, nameAndVersion string) (fingerprint string, err error) {
	operation := shared.Info("Importing " + filepath.Base(imageTar))
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()

	var meta io.ReadCloser

	meta, err = os.Open(imageTar)
	if err != nil {
		return "", err
	}
	defer meta.Close()

	image := api.ImagesPost{}

	createArgs := &lxd.ImageCreateArgs{
		MetaFile: meta,
		MetaName: filepath.Base(imageTar),
	}
	image.Filename = createArgs.MetaName

	op, err := lxdServer.CreateImage(image, createArgs)
	if err != nil {
		return "", err
	}

	err = op.Wait()
	if err != nil {
		return "", err
	}
	opAPI := op.Get()

	// Get the fingerprint
	fingerprint = opAPI.Metadata["fingerprint"].(string)

	aliasPost := api.ImageAliasesPost{}
	aliasPost.Name = nameAndVersion
	aliasPost.Target = fingerprint
	err = lxdServer.CreateImageAlias(aliasPost)

	s.Stop()

	return fingerprint, nil
}

// ExportImage downloads unit image into current directory
func ExportImage(lxdServer lxd.ImageServer, fingerprint string, name string) error {
	operation := shared.Info("Exporting " + name)
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()

	targetRootfs := name + ".root"
	dest, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dest.Close()

	destRootfs, err := os.Create(targetRootfs)
	if err != nil {
		return err
	}
	// Implicitly clean up temporary file on err
	// Defers are resolved LIFO - below ensures file closed before deletion
	defer os.Remove(targetRootfs)
	defer destRootfs.Close()

	req := lxd.ImageFileRequest{
		MetaFile:   io.WriteSeeker(dest),
		RootfsFile: io.WriteSeeker(destRootfs),
	}

	resp, err := lxdServer.GetImageFile(fingerprint, req)
	if err != nil {
		dest.Close()
		os.Remove(name)
		return err
	}

	// Truncate down to size
	if resp.RootfsSize > 0 {
		err = destRootfs.Truncate(resp.RootfsSize)
		if err != nil {
			return err
		}
	}

	err = dest.Truncate(resp.MetaSize)
	if err != nil {
		return err
	}

	dest.Close()
	destRootfs.Close()

	// Cleanup
	if resp.RootfsSize == 0 {
		err := os.Remove(targetRootfs)
		if err != nil {
			os.Remove(name)
			return err
		}
	}

	if resp.MetaName != "" {
		extension := strings.SplitN(resp.MetaName, ".", 2)[1]
		err := os.Rename(name, fmt.Sprintf("%s.%s", name, extension))
		if err != nil {
			os.Remove(name)
			return err
		}
	}

	s.Stop()
	return nil
}

func CopyImage(sourceServer lxd.InstanceServer, destServer lxd.InstanceServer, fingerprint string, alias string) error {
	operation := shared.Info(fmt.Sprintf("Copying image %q to remote", alias))
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()
	defer s.Stop()

	img, _, err := sourceServer.GetImage(fingerprint)
	if err != nil {
		return err
	}

	// Images matching the alias on the source image server must match the arch of the destination server
	destServerArch, err := GetLXDServerArch(destServer)
	if err != nil {
		return err
	}

	// Ensure no dest server conflicts
	_, err = GetImageByAlias(destServer, alias, destServerArch)
	if err == nil {
		return fmt.Errorf("image alias %q already exists on dest server", alias)
	}

	if _, _, err := destServer.GetImage(fingerprint); err == nil {
		return fmt.Errorf("image with fingerprint %q already exists on dest server", fingerprint)
	}

	args := &lxd.ImageCopyArgs{
		Mode: "relay",
	}

	op, err := destServer.CopyImage(sourceServer, *img, args)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	// Get the fingerprint
	aliasPost := api.ImageAliasesPost{}
	aliasPost.Name = alias
	aliasPost.Target = fingerprint
	err = destServer.CreateImageAlias(aliasPost)
	if err != nil {
		// Cleanup on err
		destServer.DeleteImage(fingerprint)
		return err
	}
	return nil
}

// GetFingerprintByAlias retrieves image fingerprint corresponding to provided alias
func GetFingerprintByAlias(lxdServer lxd.ImageServer, alias string, architecture string) (fingerprint string, err error) {
	if architecture == "" {
		remoteAlias, _, err := lxdServer.GetImageAlias(alias)
		if err != nil {
			return "", err
		}
		fingerprint = remoteAlias.Target
		return fingerprint, nil
	}

	// Get any matching image aliases from server and then select the correct type
	entries, err := lxdServer.GetImageAliasArchitectures("container", alias)
	if err != nil {
		return "", err
	}

	aliasEntry, ok := entries[architecture]
	if !ok {
		return "", fmt.Errorf("no image matching architecture %q found for image alias %q", architecture, alias)
	}

	fingerprint = aliasEntry.Target

	return fingerprint, nil
}

// GetImageByAlias retrieves image by name
func GetImageByAlias(lxdImageServer lxd.ImageServer, alias string, architecture string) (image *api.Image, err error) {
	imageFingerprint, err := GetFingerprintByAlias(lxdImageServer, alias, architecture)
	if err != nil {
		return nil, err
	}

	image, _, err = lxdImageServer.GetImage(imageFingerprint)
	return image, err
}

// DeleteImageFingerprint delete unit image
// lxc image delete [remote]:[name]
func DeleteImageByFingerprint(lxdServer lxd.InstanceServer, fingerprint string) error {
	op, err := lxdServer.DeleteImage(fingerprint)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}
	return nil
}

// AttachNetwork attaches unit to internal network bridge
func AttachNetwork(lxdServer lxd.InstanceServer, name string, bridge string, nic1 string, nic2 string) error {
	network, _, err := lxdServer.GetNetwork(bridge)

	if err != nil {
		return err
	}

	device := map[string]string{
		"type":    "nic",
		"nictype": "macvlan",
		"parent":  bridge,
	}

	if network.Type == "bridge" {
		device["nictype"] = "bridged"
	}

	device["name"] = nic2

	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}

	_, ok := inst.Devices[nic1]
	if ok {
		return errors.New("Device already exists: " + nic1)
	}

	inst.Devices[nic1] = device

	op, err := lxdServer.UpdateInstance(name, inst.Writable(), etag)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// ConfigDevice sets IP address
// lxc config device set [remote]:name eth0 ipv4.address
func ConfigDevice(lxdServer lxd.InstanceServer, name string, nic string, ip string) error {

	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}
	dev, ok := inst.Devices[nic]
	if !ok {
		return errors.New("device doesn't exisit")
	}

	dev["ipv4.address"] = ip
	inst.Devices[nic] = dev
	op, err := lxdServer.UpdateInstance(name, inst.Writable(), etag)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// SetConfig sets unit parameters
func SetConfig(lxdServer lxd.InstanceServer, name string, config map[string]string) error {
	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return errors.New("Error connecting to unit: " + name)
	}

	for key, value := range config {
		inst.Config[key] = value
	}

	op, err := lxdServer.UpdateInstance(name, inst.Writable(), etag)
	if err != nil {
		return errors.New("Error updating unit configuration: " + name)
	}

	err = op.Wait()
	if err != nil {
		return errors.New("Error updating unit: " + err.Error())
	}

	return nil
}

// Push ..
func Push(lxdServer lxd.InstanceServer, name string, sourcePath string, targetPath string) error {
	err := CopyDirectory(lxdServer, name, sourcePath, targetPath)
	if err != nil {
		return err
	}

	return nil
}

// CopyDirectory recursively copies a src directory to a destination.
func CopyDirectory(lxdServer lxd.InstanceServer, name string, src, dst string) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.New("Failed to read source directory: " + src)
	}
	for _, entry := range entries {
		source := filepath.Join(src, entry.Name())
		destPath := filepath.ToSlash(filepath.Join(dst, entry.Name()))

		sourcePath := filepath.FromSlash(source)

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return errors.New("Failed to get file info: " + sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createDir(lxdServer, name, destPath, 0755); err != nil {
				return errors.New("Failed to create directory: " + destPath + " : " + err.Error())
			}
			if err := CopyDirectory(lxdServer, name, sourcePath, destPath); err != nil {
				return errors.New("Failed to copy directory: " + destPath)
			}
		default:
			if err := CopyFiles(lxdServer, name, sourcePath, destPath); err != nil {
				return errors.New("Failed to copy file: " + destPath + " : " + err.Error())
			}
		}
	}
	return nil
}

// CopyFiles copies a src file to a dst file where src and dst are regular files.
func CopyFiles(lxdServer lxd.InstanceServer, name string, src, dst string) error {
	var readCloser io.ReadCloser

	fInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	mode, uid, gid := lxdshared.GetOwnerMode(fInfo)
	args := lxd.InstanceFileArgs{
		UID:  int64(uid),
		GID:  int64(gid),
		Mode: int(mode.Perm()),
	}

	f, err := os.Open(src)
	if err != nil {
		return errors.New("Failed to open source file: " + src + " : " + err.Error())
	}
	defer f.Close()

	args.Type = "file"
	args.Content = f
	readCloser = f

	contentLength, err := args.Content.Seek(0, io.SeekEnd)
	if err != nil {
		return errors.New("failed to get length of the source file")
	}

	_, err = args.Content.Seek(0, io.SeekStart)
	if err != nil {
		return errors.New("failed to get source file start")
	}

	args.Content = lxdshared.NewReadSeeker(&ioprogress.ProgressReader{
		ReadCloser: readCloser,
		Tracker: &ioprogress.ProgressTracker{
			Length: contentLength,
		},
	}, args.Content)

	log.Printf(shared.Info("Pushing %s to %s (%s)"), src, dst, args.Type)

	err = lxdServer.CreateInstanceFile(name, dst, args)
	if err != nil {
		return err
	}

	return nil
}

func createDir(lxdServer lxd.InstanceServer, name string, dir string, mode int) error {

	args := lxd.InstanceFileArgs{
		UID:  -1,
		GID:  -1,
		Mode: mode,
		Type: "directory",
	}

	log.Printf(shared.Info("Creating %s (%s)"), dir, args.Type)
	err := lxdServer.CreateInstanceFile(name, dir, args)
	if err != nil {
		return errors.New("Failed to create directory: " + dir)
	}

	return nil
}

// GetLXDInstanceServer ..
func GetLXDInstanceServer(remote Remote) (lxd.InstanceServer, error) {

	args := &lxd.ConnectionArgs{
		TLSClientKey:  remote.key,
		TLSClientCert: remote.cert,
		TLSServerCert: remote.servercert,
	}

	switch remote.Protocol {
	case "unix":
		return lxd.ConnectLXDUnix(remote.URL, args)
	case "lxd":
		return lxd.ConnectLXD(remote.URL, args)
	default:
		return nil, fmt.Errorf("unsupported protocol %q for instance server remote %q", remote.Protocol, remote.Name)
	}
}

func GetLXDImageSever(remote Remote) (lxd.ImageServer, error) {
	switch remote.Protocol {
	case "simplestreams":
		return lxd.ConnectSimpleStreams(remote.URL, nil)
	case "lxd":
		return lxd.ConnectPublicLXD(remote.URL, nil)
	default:
		return nil, fmt.Errorf("unsupported protocol %q for image sever remote %q", remote.Protocol, remote.Name)
	}
}

func GetSimplestreamsLXDSever(url string, args *lxd.ConnectionArgs) (lxd.ImageServer, error) {
	return lxd.ConnectSimpleStreams(url, args)
}

// GetLXDServerVersion retrieves server semantic version and converts to integer
func GetLXDServerVersion(lxdServer lxd.InstanceServer) (int, error) {

	serverStatus, _, err := lxdServer.GetServer()
	if err != nil {
		return -1, err
	}

	serverVersionString := strings.ReplaceAll(serverStatus.Environment.ServerVersion, ".", "")
	if len(serverVersionString) == 2 {
		serverVersionString = serverVersionString + "0"
	}

	return strconv.Atoi(serverVersionString)
}

func GetLXDServerArch(lxdServer lxd.InstanceServer) (string, error) {
	serverStatus, _, err := lxdServer.GetServer()
	if err != nil {
		return "", err
	}

	return serverStatus.Environment.KernelArchitecture, nil
}

func GetBravetoolsBridgeIP(lxdServer lxd.InstanceServer, bridgeName string) (string, error) {
	network, _, err := lxdServer.GetNetwork(bridgeName)
	if err != nil {
		return "", err
	}

	return network.Config["ipv4.address"], nil
}
