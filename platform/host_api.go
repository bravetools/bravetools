package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"io/ioutil"
	"path/filepath"

	"github.com/bravetools/bravetools/db"
	"github.com/bravetools/bravetools/shared"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
)

// Functions exposed to commands.go

// DeleteImageName deletes image by name
func (bh *BraveHost) DeleteImageName(name string) error {
	images, _ := listHostImages(bh.Remote)
	if len(images) > 0 {
		err := DeleteImageName(name, bh.Remote)
		if err != nil {
			return errors.New("Image: " + err.Error())
		}
	}

	return nil
}

// DeleteImage delete image by fingerprint
func (bh *BraveHost) DeleteImage(fingerprint string) error {
	err := DeleteImage(fingerprint, bh.Remote)
	if err != nil {
		return errors.New("Failed to delete image: " + err.Error())
	}

	return nil
}

// DeleteHostImages removes all LXC images from host
func (bh *BraveHost) DeleteHostImages() error {
	err := deleteHostImages(bh.Remote)
	if err != nil {
		return err
	}
	return nil
}

// AddRemote sets connection to Brave platform
func (bh *BraveHost) AddRemote() error {
	url := "https://" + bh.Settings.BackendSettings.Resources.IP + ":8443"
	bh.Remote.remoteURL = url

	err := RemoveRemote(bh.Settings.Name)
	if err != nil {
		fmt.Println("No Brave host. Continue adding a new host ..")
	}
	err = AddRemote(bh)
	if err != nil {
		return errors.New("Failed to add remote host: " + err.Error())
	}

	if err != nil {
		log.Fatal("Failed to access user home directory: ", err.Error())
	}

	return nil
}

// ImportLocalImage import tarball into local images folder
func (bh *BraveHost) ImportLocalImage(name string) error {
	home, _ := os.UserHomeDir()
	imagePath := home + shared.ImageStore
	hashFile := imagePath + name + ".md5"

	_, err := os.Stat(home + shared.ImageStore + name)
	if !os.IsNotExist(err) {
		return errors.New("Image " + name + " already exists in local image storage")
	}

	err = shared.CopyFile(name, imagePath+name)
	if err != nil {
		return errors.New("Failed to copy image archive to local image storage: " + err.Error())
	}

	imageHash, err := shared.FileHash(name)
	if err != nil {
		return errors.New("Failed to generate image hash: " + err.Error())
	}

	fmt.Println(imageHash)

	// Write image hash to a file
	f, err := os.Create(hashFile)
	if err != nil {
		return errors.New(err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(imageHash)
	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}

// ListLocalImages reads list of files in images folder
func (bh *BraveHost) ListLocalImages() error {
	home, _ := os.UserHomeDir()
	imagePath := home + shared.ImageStore

	// We're only interested in images and not MD5 checksums
	images, err := shared.WalkMatch(imagePath, "*.tar.gz")
	if err != nil {
		return errors.New("Failed to access images folder: " + err.Error())
	}

	if len(images) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Image", "Created", "Size", "Hash"})

		for _, i := range images {
			fi, err := os.Stat(i)
			if strings.Index(fi.Name(), ".") != 0 {
				if err != nil {
					return errors.New("Failed to get image size: " + err.Error())
				}

				name := strings.Split(fi.Name(), ".tar.gz")[0]

				size := fi.Size()

				created := int(time.Since(fi.ModTime()).Hours() / 24)
				var timeUnit string
				if created > 1 {
					timeUnit = strconv.Itoa(created) + " days go"
				} else if created == 1 {
					timeUnit = strconv.Itoa(created) + " day ago"
				} else {
					timeUnit = "just now"
				}

				localImageFile := home + shared.ImageStore + filepath.Base(fi.Name())
				hashFileName := localImageFile + ".md5"

				hash, err := ioutil.ReadFile(hashFileName)
				if err != nil {
					if os.IsNotExist(err) {

						imageHash, err := shared.FileHash(localImageFile)
						if err != nil {
							return errors.New("Failed to generate image hash: " + err.Error())
						}

						f, err := os.Create(hashFileName)
						if err != nil {
							return errors.New(err.Error())
						}
						defer f.Close()

						_, err = f.WriteString(imageHash)
						if err != nil {
							return errors.New(err.Error())
						}

						hash, err = ioutil.ReadFile(hashFileName)
					} else {
						return errors.New("Couldn't load image hash: " + err.Error())
					}
				}

				hashString := string(hash)
				hashString = strings.TrimRight(hashString, "\r\n")

				r := []string{name, timeUnit, shared.FormatByteCountSI(size), hashString}
				table.Append(r)
			}
		}

		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
		table.Render()

	} else {
		fmt.Println("No local images")
	}

	return nil
}

// DeleteLocalImage deletes a local image
func (bh *BraveHost) DeleteLocalImage(name string) error {
	home, _ := os.UserHomeDir()
	imagePath := home + shared.ImageStore
	imageName := imagePath + name + ".tar.gz"
	imageHash := imageName + ".md5"

	err := os.Remove(imageName)
	if err != nil {
		return err
	}

	err = os.Remove(imageHash)
	if err != nil {
		return err
	}

	return nil
}

// HostInfo returns useful information about brave host
func (bh *BraveHost) HostInfo(backend Backend, short bool) error {
	info, err := backend.Info()
	if err != nil {
		return errors.New("Failed to connect to host: " + err.Error())
	}

	if short {
		fmt.Println(info.IPv4)
		return nil
	}

	if info.State == "Stopped" {
		return errors.New("Cannot connect to Bravetools remote, ensure it is up and running")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "State", "IPv4", "Disk", "Memory", "CPU"})

	r := []string{info.Name, info.State, info.IPv4,
		info.Disk[0] + " of " + info.Disk[1],
		info.Memory[0] + " of " + info.Memory[1], info.CPU}

	table.Append(r)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.Render()

	return nil
}

// ListUnits prints all LXD containers on remote host
func (bh *BraveHost) ListUnits(backend Backend) error {
	info, err := backend.Info()
	if err != nil {
		return err
	}

	if info.State == "Stopped" {
		return errors.New("Cannot connect to Bravetools remote, ensure it is up and running")
	}

	units, err := listHostUnits(bh.Remote)
	if err != nil {
		return errors.New("Failed to list units: " + err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "IPv4", "Disk", "Proxy"})
	for _, u := range units {
		name := u.Name
		status := u.Status
		address := u.Address

		disk := ""
		if u.Disk.Name != "" {
			disk = u.Disk.Name + ":" + u.Disk.Source + "->" + u.Disk.Path
		}

		r := []string{name, status, u.NIC.Name + ":" + address, disk, u.Proxy.Name}
		table.Append(r)
	}
	table.SetRowLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.Render()

	return nil
}

// UmountDirectory ..
func (bh *BraveHost) UmountDirectory(unit string, target string) error {
	backend := bh.Settings.BackendSettings.Type

	switch backend {
	case "multipass":
		path, err := DeleteDevice(unit, target, bh.Remote)
		if err != nil {
			return errors.New("Failed to umount " + target + ": " + err.Error())
		}

		cmd := fmt.Sprintf(`if [ -d "%s" ]; then echo "exists"; else echo "none"; fi`, path)
		output, err := shared.ExecCommandWReturn("multipass",
			"exec",
			bh.Settings.Name,
			"--", "bash", "-c",
			cmd)
		if err != nil {
			return errors.New("Could not check directory: " + err.Error())
		}
		output = strings.Trim(output, "\n")

		if output == "exists" {
			err = shared.ExecCommand("multipass",
				"umount",
				bh.Settings.Name+":"+path)
			if err != nil {
				return err
			}
		}

	case "lxd":
		_, err := DeleteDevice(unit, target, bh.Remote)
		if err != nil {
			return errors.New("Failed to umount " + target + ": " + err.Error())
		}
	}

	volume, _ := GetVolume(bh.Settings.StoragePool.Name, bh.Remote)
	if len(volume.UsedBy) == 0 {
		DeleteVolume(bh.Settings.StoragePool.Name, volume, bh.Remote)

		return nil
	}

	return nil
}

// MountDirectory ..
func (bh *BraveHost) MountDirectory(source string, destUnit string, destPath string) error {
	backend := bh.Settings.BackendSettings.Type
	var sourceUnit string
	var sourcePath string

	sourceSlice := strings.SplitN(source, ":", -1)
	if len(sourceSlice) > 2 {
		return errors.New("Failed to parse source " + source + "Accepted form [UNIT:]<path>")
	} else if len(sourceSlice) == 2 {
		sourceUnit = sourceSlice[0]
		sourcePath = sourceSlice[1]
	} else if len(sourceSlice) == 1 {
		sourceUnit = ""
		sourcePath = source
	}

	sharedDirectory := filepath.Base(sourcePath)

	switch backend {
	case "multipass":

		if sourceUnit == "" {
			err := shared.ExecCommand("multipass",
				"mount",
				sourcePath,
				bh.Settings.Name+":/home/ubuntu/volumes/"+sharedDirectory)
			if err != nil {
				return errors.New("Failed to initialize mount on host :" + err.Error())
			}

			err = MountDirectory(filepath.Join("/home/ubuntu", "volumes", sharedDirectory), destUnit, destPath, bh.Remote)
			if err != nil {
				return errors.New("Failed to mount " + sourcePath + " to " + destUnit + ":" + destPath + " : " + err.Error())
			}
		} else {
			err := createSharedVolume(bh.Settings.StoragePool.Name,
				sharedDirectory,
				sourceUnit,
				destUnit,
				destPath,
				bh)
			if err != nil {
				return err
			}
		}
	case "lxd":
		if sourceUnit == "" {
			err := MountDirectory(sourcePath, destUnit, destPath, bh.Remote)
			if err != nil {
				return errors.New("Failed to mount " + source + " to " + destUnit + ":" + destPath + " : " + err.Error())
			}
		} else {
			err := createSharedVolume(bh.Settings.StoragePool.Name,
				sharedDirectory,
				sourceUnit,
				destUnit,
				destPath,
				bh)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteUnit ..
func (bh *BraveHost) DeleteUnit(name string) error {
	var unitNames []string

	unitList, err := listHostUnits(bh.Remote)
	if err != nil {
		return errors.New("Failed to list existing units: " + err.Error())
	}

	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	if !shared.StringInSlice(name, unitNames) {
		return errors.New("Unit " + name + " does not exist")
	}

	err = Delete(name, bh.Remote)
	if err != nil {
		return errors.New("Failed to delete unit: " + err.Error())
	}

	// Deleting unit from databse

	userHome, err := os.UserHomeDir()
	if err != nil {
		return errors.New("Failed to get home directory")
	}
	dbPath := path.Join(userHome, shared.BraveDB)
	database := db.OpenDB(dbPath)

	err = db.DeleteUnitDB(database, name)
	if err != nil {
		return errors.New("Failed to delete unit from database. Name: " + name + " Error: " + err.Error())
	}

	fmt.Println("Unit deleted: ", name)
	return nil
}

// BuildUnit creates unit based on Bravefile
func (bh *BraveHost) BuildUnit(bravefile *shared.Bravefile) error {
	var err error
	var fingerprint string
	var unitNames []string

	if strings.ContainsAny(bravefile.PlatformService.Name, "/_. !@£$%^&*(){}:;`~,?") {
		return errors.New("Image names should not contain special characters")
	}

	if bravefile.PlatformService.Name == "" {
		return errors.New("Service Name is empty")
	}

	unitList, err := listHostUnits(bh.Remote)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	unitExists := shared.StringInSlice(bravefile.PlatformService.Name, unitNames)
	if unitExists {
		return errors.New("Unit " + bravefile.PlatformService.Name + " already exists on host")
	}

	images, err := listHostImages(bh.Remote)
	if err != nil {
		log.Fatal(err)
	}

	if len(images) > 0 {
		err = deleteHostImages(bh.Remote)
		if err != nil {
			return err
		}
	}

	processInterruptHandler(fingerprint, bravefile, bh)

	switch bravefile.Base.Location {
	case "public":
		err = importLXD(bravefile, bh.Remote)
		if err != nil {
			log.Fatal(err)
		}

		err = Start(bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			log.Fatal(err)
		}
	case "github":
		err = importGitHub(bravefile, bh)
		if err != nil {
			log.Fatal(err)
		}

		err = Start(bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			log.Fatal(err)
		}
	default:
		if bravefile.Base.Location == "local" {
			err = importLocal(bravefile, bh.Remote)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	time.Sleep(10 * time.Second)
	pMan := bravefile.SystemPackages.Manager

	switch pMan {
	case "apk":
		_, err := Exec(bravefile.PlatformService.Name, []string{"apk", "update"}, bh.Remote)
		if err != nil {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New("Failed to update repositories: " + err.Error())
		}

		args := []string{"apk", "--no-cache", "add"}
		for _, p := range bravefile.SystemPackages.System {
			args = append(args, p)
		}
		status, err := Exec(bravefile.PlatformService.Name, args, bh.Remote)
		if err != nil {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New("Failed to install packages: " + err.Error())
		}
		if status > 0 {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New(shared.Fatal("Failed to install packages"))
		}

	case "apt":
		_, err := Exec(bravefile.PlatformService.Name, []string{"apt", "update"}, bh.Remote)
		if err != nil {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New("Failed to update repositories: " + err.Error())
		}

		args := []string{"apt", "install"}
		for _, p := range bravefile.SystemPackages.System {
			args = append(args, p)
		}
		args = append(args, "--yes")
		status, err := Exec(bravefile.PlatformService.Name, args, bh.Remote)
		if err != nil {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New("Failed to install packages: " + err.Error())
		}
		if status > 0 {
			DeleteImage(fingerprint, bh.Remote)
			Delete(bravefile.PlatformService.Name, bh.Remote)
			return errors.New(shared.Fatal("Failed to install packages"))
		}
	}

	// Go through "Copy" section
	err = bravefileCopy(bravefile.Copy, bravefile.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)
		return err
	}

	// Go through "Run" section
	status, err := bravefileRun(bravefile.Run, bravefile.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)
		return errors.New("Failed to execute command: " + err.Error())
	}
	if status > 0 {
		DeleteImage(fingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)
		return errors.New(shared.Fatal("Failed to execute command"))
	}

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	fmt.Println("Publishing image " + bravefile.PlatformService.Name)

	var unitFingerprint string
	unitFingerprint, err = Publish(bravefile.PlatformService.Name, bravefile.PlatformService.Version, bh.Remote)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)
		return errors.New("Failed to publish image: " + err.Error())
	}

	fmt.Println("Exporting image " + bravefile.PlatformService.Name)
	err = ExportImage(unitFingerprint, bravefile.PlatformService.Name+"-"+bravefile.PlatformService.Version, bh.Remote)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		DeleteImage(unitFingerprint, bh.Remote)
		Delete(bravefile.PlatformService.Name, bh.Remote)
		return errors.New("Failed to export image: " + err.Error())
	}

	DeleteImage(fingerprint, bh.Remote)
	DeleteImage(unitFingerprint, bh.Remote)
	Delete(bravefile.PlatformService.Name, bh.Remote)

	home, _ := os.UserHomeDir()
	localImageFile := bravefile.PlatformService.Name + "-" + bravefile.PlatformService.Version + ".tar.gz"
	localHashFile := localImageFile + ".md5"

	imageHash, err := shared.FileHash(localImageFile)
	if err != nil {
		return errors.New("Failed to generate image hash: " + err.Error())
	}

	fmt.Println(imageHash)

	// Write image hash to a file
	f, err := os.Create(localHashFile)
	if err != nil {
		return errors.New(err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(imageHash)
	if err != nil {
		return errors.New(err.Error())
	}
	f.Close()

	err = shared.CopyFile(localImageFile, home+shared.ImageStore+localImageFile)
	if err != nil {
		return errors.New("Failed to copy image archive to local storage: " + err.Error())
	}

	err = shared.CopyFile(localHashFile, home+shared.ImageStore+localHashFile)
	if err != nil {
		return errors.New("Failed to copy images hash into local storage: " + err.Error())
	}

	err = os.Remove(localImageFile)
	if err != nil {
		return errors.New("Failed to clean up image archive: " + err.Error())
	}

	err = os.Remove(localHashFile)
	if err != nil {
		return errors.New("Failed to clean up image hash: " + err.Error())
	}

	return nil
}

// PublishUnit publishes unit to image
func (bh *BraveHost) PublishUnit(name string, backend Backend) error {
	_, err := backend.Info()
	if err != nil {
		return errors.New("Failed to get host info: " + err.Error())
	}

	timestamp := time.Now()

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	fmt.Println("Publishing unit ...")

	var unitFingerprint string
	unitFingerprint, err = Publish(name, timestamp.Format("20060102150405"), bh.Remote)
	if err != nil {
		DeleteImage(unitFingerprint, bh.Remote)
		return errors.New("Failed to publish image: " + err.Error())
	}

	fmt.Println("Exporting archive ...")
	err = ExportImage(unitFingerprint, name+"-"+timestamp.Format("20060102150405"), bh.Remote)
	if err != nil {
		DeleteImage(unitFingerprint, bh.Remote)
		return errors.New("Failed to export unit: " + err.Error())
	}

	fmt.Println("Cleaning ...")
	DeleteImage(unitFingerprint, bh.Remote)

	return nil
}

// StopUnit stops unit using name
func (bh *BraveHost) StopUnit(name string, backend Backend) error {
	info, err := backend.Info()
	if err != nil {
		return errors.New("Failed to get host info: " + err.Error())
	}
	if strings.ToLower(info.State) == "stopped" {
		return errors.New("Backend is stopped")
	}
	fmt.Print("Stopping unit: ", name)
	err = Stop(name, bh.Remote)
	if err != nil {
		return errors.New("Failed to stop unit: " + err.Error())
	}

	return nil
}

// StartUnit restarts unit if running and starts if stopped.
func (bh *BraveHost) StartUnit(name string, backend Backend) error {
	info, err := backend.Info()
	if err != nil {
		return errors.New("Failed to get host info: " + err.Error())
	}
	if strings.ToLower(info.State) == "stopped" {
		return errors.New("Backend is stopped")
	}
	fmt.Print("Starting unit: ", name)
	err = Start(name, bh.Remote)
	if err != nil {
		return errors.New("Failed to start unit: " + err.Error())
	}

	return nil
}

// InitUnit starts unit from supplied image
func (bh *BraveHost) InitUnit(backend Backend, unitParams *shared.Bravefile) error {
	var err error
	var fingerprint string
	var unitNames []string

	homeDir, _ := os.UserHomeDir()
	if unitParams.PlatformService.Image == "" {
		return errors.New("Unit image name cannot be empty")
	}
	image := homeDir + shared.ImageStore + unitParams.PlatformService.Image + ".tar.gz"

	fi, err := os.Stat(image)
	if err != nil {
		return err
	}

	requestedImageSize := fi.Size()

	// Resource checks
	info, err := backend.Info()

	usedDiskSize, err := shared.SizeCountToInt(info.Disk[0])
	if err != nil {
		return err
	}
	totalDiskSize, err := shared.SizeCountToInt(info.Disk[1])
	if err != nil {
		return err
	}

	if requestedImageSize*5 > (totalDiskSize - usedDiskSize) {
		return errors.New("Requested unit size exceeds available disk space on bravetools host")
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
				return errors.New("Invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
			}
			hostPorts = append(hostPorts, ps[1])
		}
	}
	err = shared.TCPPortStatus(hostIP, hostPorts)
	if err != nil {
		return err
	}

	// Unit Checks
	unitList, err := listHostUnits(bh.Remote)
	if err != nil {
		return err
	}

	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	unitExists := shared.StringInSlice(unitParams.PlatformService.Name, unitNames)
	if unitExists {
		return errors.New("Unit " + unitParams.PlatformService.Name + " already exists on host")
	}

	fingerprint, err = ImportImage(image, unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		return errors.New("Failed to import image: " + err.Error())
	}

	err = LaunchFromImage(unitParams.PlatformService.Name, unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		return errors.New("Failed to launch unit: " + err.Error())
	}

	err = AttachNetwork(unitParams.PlatformService.Name, "lxdbr0", "eth0", "eth0", bh.Remote)
	if err != nil {
		return errors.New("Failed to attach network: " + err.Error())
	}

	err = ConfigDevice(unitParams.PlatformService.Name, "eth0", unitParams.PlatformService.IP, bh.Remote)
	if err != nil {
		return errors.New("Failed to set IP: " + err.Error())
	}

	err = Stop(unitParams.PlatformService.Name, bh.Remote)
	err = Start(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		return errors.New("Failed to restart unit: " + err.Error())
	}

	config := map[string]string{
		"limits.cpu":       unitParams.PlatformService.Resources.CPU,
		"limits.memory":    unitParams.PlatformService.Resources.RAM,
		"security.nesting": "false",
		"nvidia.runtime":   "false",
	}

	if unitParams.PlatformService.Docker == "yes" {
		config["security.nesting"] = "true"
	}

	if unitParams.PlatformService.Resources.GPU == "yes" {
		config["nvidia.runtime"] = "true"
		device := map[string]string{"type": "gpu"}
		err = AddDevice(unitParams.PlatformService.Name, "gpu", device, bh.Remote)
	}

	err = SetConfig(unitParams.PlatformService.Name, config, bh.Remote)
	if err != nil {
		return errors.New("Error configuring unit: " + err.Error())
	}

	err = Stop(unitParams.PlatformService.Name, bh.Remote)
	err = Start(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		return errors.New("Failed to restart unit: " + err.Error())
	}

	fmt.Println("Service started: ", unitParams.PlatformService.Name)

	ports = unitParams.PlatformService.Ports
	if len(ports) > 0 {
		for _, p := range ports {
			ps := strings.Split(p, ":")
			if len(ps) < 2 {
				DeleteImage(fingerprint, bh.Remote)
				Delete(unitParams.PlatformService.Name, bh.Remote)
				return errors.New("Invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
			}

			err := addIPRules(unitParams.PlatformService.Name, ps[1], ps[0], bh)
			if err != nil {
				err = Delete(unitParams.PlatformService.Name, bh.Remote)
				if err != nil {
					return errors.New("Failed to delete unit: " + err.Error())
				}
				log.Fatal(err)
			}
		}
	}

	// Add unit into database

	var braveUnit db.BraveUnit
	userHome, err := os.UserHomeDir()
	if err != nil {
		return errors.New("Failed to get home directory")
	}
	dbPath := path.Join(userHome, shared.BraveDB)

	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {

		err = db.InitDB(dbPath)

		if err != nil {
			DeleteImage(fingerprint, bh.Remote)
			Delete(unitParams.PlatformService.Name, bh.Remote)
			return errors.New("Failed to initialize database: " + err.Error())
		}
	}

	log.Println("Connecting to database")
	database := db.OpenDB(dbPath)

	uuid, _ := uuid.NewUUID()
	braveUnit.UID = uuid.String()
	braveUnit.Name = unitParams.PlatformService.Name
	braveUnit.Date = time.Now().String()

	var unitData db.UnitData
	unitData.CPU, _ = strconv.Atoi(unitParams.PlatformService.Resources.CPU)
	unitData.RAM = unitParams.PlatformService.Resources.RAM
	unitData.IP = unitParams.PlatformService.IP
	unitData.Image = unitParams.Base.Image

	data, err := json.Marshal(unitData)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		Delete(unitParams.PlatformService.Name, bh.Remote)
		return errors.New("Failed to serialize unit data")
	}
	braveUnit.Data = data
	log.Println("Inserting unit")
	_, err = db.InsertUnitDB(database, braveUnit)
	if err != nil {
		DeleteImage(fingerprint, bh.Remote)
		Delete(unitParams.PlatformService.Name, bh.Remote)
		return errors.New("Failed to insert unit to database: " + err.Error())
	}

	return nil
}

// Postdeploy copy files and run commands on running service
func (bh *BraveHost) Postdeploy(bravefile *shared.Bravefile) (err error) {

	if bravefile.PlatformService.Postdeploy.Copy != nil {
		err = bravefileCopy(bravefile.PlatformService.Postdeploy.Copy, bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			return err
		}
	}

	if bravefile.PlatformService.Postdeploy.Run != nil {
		_, err = bravefileRun(bravefile.PlatformService.Postdeploy.Run, bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			return err
		}
	}

	return nil
}
