package platform

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/beringresearch/bravetools/shared"

	pem "encoding/pem"

	lxd "github.com/lxc/lxd/client"
	lxdshared "github.com/lxc/lxd/shared"
	api "github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/ioprogress"
)

// DeleteStoragePool ..
func DeleteStoragePool(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	err := lxdServer.DeleteStoragePool(name)
	if err != nil {
		return errors.New("Failed to delete storage pool: " + err.Error())
	}

	return nil
}

// SetActiveStoragePool pool assigns a profile with default storage
func SetActiveStoragePool(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	profile, etag, err := lxdServer.GetProfile("brave")
	if err != nil {
		return errors.New("Unable to load brave profile: " + err.Error())
	}

	device := map[string]string{}

	device["type"] = "disk"
	device["path"] = "/"
	device["pool"] = name

	profile.Devices["root"] = device

	err = lxdServer.UpdateProfile("brave", profile.Writable(), etag)
	if err != nil {
		return errors.New("Failed to update brave profile with storage pool configuration: " + err.Error())
	}

	return nil
}

// CreateStoragePool creates a new storage pool
func CreateStoragePool(name string, size string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	req := api.StoragePoolsPost{
		Name:   name,
		Driver: "zfs",
	}

	req.Config = map[string]string{
		"size": size,
	}

	err := lxdServer.CreateStoragePool(req)
	if err != nil {
		return errors.New("Failed to create storage pool: " + err.Error())
	}

	return nil
}

// AddRemote adds remote LXC host
func AddRemote(braveHost *BraveHost) error {
	var err error
	userHome, _ := os.UserHomeDir()
	certf := userHome + shared.BraveClientCert
	keyf := userHome + shared.BraveClientKey

	// Generate client certificates
	err = lxdshared.FindOrGenCert(certf, keyf, true, false)

	// Check if the system CA worked for the TLS connection
	var certificate *x509.Certificate
	certificate, err = lxdshared.GetRemoteCertificate(braveHost.Remote.remoteURL, "")
	if err != nil {
		return err
	}

	// Handle certificate prompt
	if certificate != nil {
		digest := lxdshared.CertFingerprint(certificate)
		fmt.Printf(("Certificate fingerprint: %s")+"\n", digest)
	}

	dnam := userHome + "/.bravetools/" + "servercerts"
	err = os.MkdirAll(dnam, 0750)
	if err != nil {
		return fmt.Errorf(("Could not create server cert dir"))
	}

	certf = fmt.Sprintf("%s/%s.crt", dnam, braveHost.Settings.Name)
	certOut, err := os.Create(certf)
	if err != nil {
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	certOut.Close()

	req := api.CertificatesPost{
		Password: braveHost.Settings.Trust,
	}
	req.Type = "client"

	keyPath := userHome + shared.BraveClientKey
	certPath := userHome + shared.BraveClientCert
	key, _ := loadKey(keyPath)
	cert, _ := loadCert(certPath)

	lxdServer := GetLXDServer(key, cert, braveHost.Remote.remoteURL)

	err = lxdServer.CreateCertificate(req)
	if err != nil {
		return err
	}

	return nil
}

// RemoveRemote removes remote LXC host
func RemoveRemote(name string) error {
	userHome, _ := os.UserHomeDir()
	certf := userHome + shared.BraveClientCert
	keyf := userHome + shared.BraveClientKey
	certs := userHome + "/.bravetools/" + "servercerts/" + name + ".crt"
	err := os.Remove(certf)
	if err != nil {
		return err
	}
	err = os.Remove(keyf)
	if err != nil {
		return err
	}
	err = os.Remove(certs)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDevice unmounts a disk
func DeleteDevice(name string, target string, remote Remote) (string, error) {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return "", err
	}

	devname := target
	device, ok := inst.Devices[devname]
	if !ok {
		return "", errors.New("Device " + devname + " doesn't exist")
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

// AddDevice adds an external device to
func AddDevice(ct string, devname string, devices map[string]string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	inst, etag, err := lxdServer.GetInstance(ct)
	if err != nil {
		return errors.New("Error accessing unit: " + ct)
	}

	inst.Devices[devname] = devices

	op, err := lxdServer.UpdateInstance(ct, inst.Writable(), etag)
	if err != nil {
		return errors.New("Errors updating unit configuration: " + ct)
	}

	err = op.Wait()
	if err != nil {
		return errors.New("Error updating unit " + ct + " Error: " + err.Error())
	}

	return nil

}

// MountDirectory mounts local directory to unit
func MountDirectory(name string, source string, destination string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}

	devname := "disk" + shared.RandomSequence(2)
	_, ok := inst.Devices[devname]
	if ok {
		return errors.New("Unable to mount directory as duplicate device found")
	}

	device := map[string]string{}
	device["type"] = "disk"
	device["source"] = source
	device["path"] = destination

	inst.Devices[devname] = device

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

// GetImages returns all images from host
func GetImages(remote Remote) ([]api.Image, error) {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	images, err := lxdServer.GetImages()
	if err != nil {
		return nil, err
	}

	return images, nil
}

// GetUnits returns all running units
func GetUnits(remote Remote) ([]api.InstanceFull, error) {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	units, err := lxdServer.GetInstancesFull(api.InstanceTypeAny)
	if err != nil {
		return nil, err
	}

	return units, nil
}

// LaunchFromImage creates new unit based on image
func LaunchFromImage(image string, name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	req := api.ContainersPost{
		Name: name,
	}

	alias, _, err := lxdServer.GetImageAlias(image)
	if err != nil {
		return err
	}
	req.Source.Alias = name

	//TODO: obtain profile from settings
	req.Profiles = []string{"brave"}

	image = alias.Target
	imgInfo, _, err := lxdServer.GetImage(image)
	if err != nil {
		return err
	}

	//TODO: method of InstanceServer requires itself
	op, err := lxdServer.CreateContainerFromImage(lxdServer, *imgInfo, req)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	fmt.Println("Unit launched: ", name)
	return nil
}

// Launch starts a new unit based on standard image from linuxcontainers.org
// Alias: "ubuntu/bionic/amd64"
// Alias: "alpine/3.9/amd64"
func Launch(name string, alias string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

	req := api.ContainersPost{
		Name: name,
		Source: api.ContainerSource{
			Type:     "image",
			Protocol: "simplestreams",
			Server:   "https://images.linuxcontainers.org/",
			Alias:    alias,
		},
	}

	//TODO: obtain profile from settings
	req.Profiles = []string{"brave"}

	fmt.Println("Creating " + name)
	op, err := lxdServer.CreateContainer(req)
	if err != nil {
		return errors.New("Failed to create unit: " + err.Error())
	}

	err = op.Wait()
	if err != nil {
		return errors.New("Error waiting: " + err.Error())
	}

	fmt.Println("Unit launched: ", name, alias)
	return nil
}

// Exec runs command inside unit
func Exec(name string, command []string, remote Remote) (int, error) {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	fmt.Println(shared.Info("["+name+"] "+"RUN: "), command)

	req := api.InstanceExecPost{
		Command:      command,
		WaitForWS:    true,
		RecordOutput: true,
		Interactive:  false,
	}

	args := lxd.InstanceExecArgs{
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Control:  nil, // terminal non-interactive
		DataDone: make(chan bool),
	}

	op, err := lxdServer.ExecInstance(name, req, &args)

	if err != nil {
		return 1, errors.New("Error getting current state: " + err.Error())
	}

	err = op.Wait()
	if err != nil {
		return 1, errors.New("Error executing command: " + err.Error())
	}
	opAPI := op.Get()

	<-args.DataDone
	status := int(opAPI.Metadata["return"].(float64))

	return status, nil
}

// Delete unit
func Delete(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	unit, _, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}

	devices := []string{}
	for key, value := range unit.Devices {
		if value["type"] == "disk" {
			devices = append(devices, key)
		}
	}
	if len(devices) > 0 {
		return errors.New("Cannot delete unit " + name + " due to mounted disks. Umount them and try again")
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
			return errors.New("Stopping the instance failed: " + err.Error())
		}
	}

	op, err := lxdServer.DeleteContainer(name)
	if err != nil {
		return errors.New("Fail to delete unit: " + err.Error())
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Start unit
func Start(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
func Stop(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
// lxc publish -f [remote]:[name] [remote]: --alias [name-version]
func Publish(name string, version string, remote Remote) (fingerprint string, err error) {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
	aliasPost.Name = name + "-" + version
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

	return fingerprint, nil
}

// SymlinkPush  copies a symlink into unit
func SymlinkPush(name string, sourceFile string, targetPath string, remote Remote) error {
	var readCloser io.ReadCloser
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)

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
func FilePush(name string, sourceFile string, targetPath string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	var readCloser io.ReadCloser
	fInfo, err := os.Stat(sourceFile)

	if err != nil {
		return errors.New("Unable to read file " + sourceFile + ": " + err.Error())
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

	fmt.Printf(shared.Info("Pushing %s to %s (%s)\n"), sourceFile, targetPath, args.Type)

	_, targetFile := filepath.Split(sourceFile)

	err = lxdServer.CreateInstanceFile(name, filepath.Join(targetPath, targetFile), args)
	if err != nil {
		return err
	}

	return nil
}

// ImportImage imports image from current directory
func ImportImage(imageTar string, nameAndVersion string, remote Remote) (fingerprint string, err error) {
	fmt.Println("Importing " + filepath.Base(imageTar))
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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

	return fingerprint, nil
}

// ExportImage downloads unit image into current directory
func ExportImage(fingerprint string, name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
	defer destRootfs.Close()

	req := lxd.ImageFileRequest{
		MetaFile:   io.WriteSeeker(dest),
		RootfsFile: io.WriteSeeker(destRootfs),
	}

	resp, err := lxdServer.GetImageFile(fingerprint, req)
	if err != nil {
		os.Remove(name)
		os.Remove(targetRootfs)
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

	// Cleanup
	if resp.RootfsSize == 0 {
		err := os.Remove(targetRootfs)
		if err != nil {
			os.Remove(name)
			os.Remove(targetRootfs)
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

	return nil
}

// DeleteImageName delete unit image by name
func DeleteImageName(name string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	err := lxdServer.DeleteImageAlias(name)
	if err != nil {
		return err
	}

	fmt.Println(name)
	return nil
}

// DeleteImage delete unit image
// lxc image delete [remote]:[name]
func DeleteImage(fingerprint string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
// lxc network attach [remote]lxdbr0 [name] eth0 eth0
func AttachNetwork(name string, bridge string, nic1 string, nic2 string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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

	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// ConfigDevice sets IP address
// lxc config device set [remote]:name eth0 ipv4.address
func ConfigDevice(name string, nic string, ip string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	inst, etag, err := lxdServer.GetInstance(name)
	if err != nil {
		return err
	}
	dev, ok := inst.Devices[nic]
	if !ok {
		return errors.New("The device doesn't exisit")
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
func SetConfig(name string, config map[string]string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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
func Push(name string, sourcePath string, targetPath string, remote Remote) error {
	err := CopyDirectory(name, sourcePath, targetPath, remote)
	if err != nil {
		return err
	}

	return nil
}

// CopyDirectory recursively copies a src directory to a destination.
func CopyDirectory(name string, src, dst string, remote Remote) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.New("Failed to read source directory: " + src)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dst, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return errors.New("Failed to get file info: " + sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createDir(name, destPath, 0755, remote); err != nil {
				return errors.New("Failed to create directory: " + destPath + " : " + err.Error())
			}
			if err := CopyDirectory(name, sourcePath, destPath, remote); err != nil {
				return errors.New("Failed to copy directory: " + destPath)
			}
		default:
			if err := CopyFiles(name, sourcePath, destPath, remote); err != nil {
				return errors.New("Failed to copy file: " + destPath + " : " + err.Error())
			}
		}
	}
	return nil
}

// CopyFiles copies a src file to a dst file where src and dst are regular files.
func CopyFiles(name string, src, dst string, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
	var readCloser io.ReadCloser

	fInfo, err := os.Stat(src)

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
		return errors.New("Failed to get lenght of the source file")
	}

	_, err = args.Content.Seek(0, io.SeekStart)
	if err != nil {
		return errors.New("Failed to get source file start")
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

func createDir(name string, dir string, mode int, remote Remote) error {
	lxdServer := GetLXDServer(remote.key, remote.cert, remote.remoteURL)
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

// GetLXDServer ..
func GetLXDServer(key string, cert string, url string) lxd.InstanceServer {
	args := lxd.ConnectionArgs{}
	args.TLSClientKey = key
	args.TLSClientCert = cert
	args.InsecureSkipVerify = true

	server, err := lxd.ConnectLXD(url, &args)
	if err != nil {
		log.Fatal(errors.New("Failed to connect to LXD: " + err.Error()))
		return nil
	}

	return server
}
