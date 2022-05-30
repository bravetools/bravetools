package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"runtime"
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
func (bh *BraveHost) DeleteImageByName(name string) error {
	images, _ := listHostImages(bh.Remote)
	if len(images) > 0 {
		err := DeleteImageByName(name, bh.Remote)
		if err != nil {
			return errors.New("image: " + err.Error())
		}
	}

	return nil
}

// DeleteImage delete image by fingerprint
func (bh *BraveHost) DeleteImageByFingerprint(fingerprint string) error {
	err := DeleteImageByFingerprint(fingerprint, bh.Remote)
	if err != nil {
		return errors.New("failed to delete image: " + err.Error())
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
		fmt.Println("no Brave host. Continue adding a new host ..")
	}
	err = AddRemote(bh)
	if err != nil {
		return errors.New("failed to add remote host: " + err.Error())
	}

	if err != nil {
		log.Fatal("failed to access user home directory: ", err.Error())
	}

	return nil
}

// ImportLocalImage import tarball into local images folder
func (bh *BraveHost) ImportLocalImage(sourcePath string) error {
	home, _ := os.UserHomeDir()

	_, imageName := filepath.Split(sourcePath)

	imagePath := home + shared.ImageStore
	hashFile := imagePath + imageName + ".md5"

	_, err := os.Stat(home + shared.ImageStore + imageName)
	if !os.IsNotExist(err) {
		return errors.New("image " + imageName + " already exists in local image store")
	}

	err = shared.CopyFile(sourcePath, imagePath+imageName)
	if err != nil {
		return errors.New("failed to copy image archive to local image store: " + err.Error())
	}

	imageHash, err := shared.FileHash(sourcePath)
	if err != nil {
		return errors.New("failed to generate image hash: " + err.Error())
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
		return errors.New("failed to access images folder: " + err.Error())
	}

	if len(images) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Image", "Created", "Size", "Hash"})

		for _, i := range images {
			fi, err := os.Stat(i)
			if strings.Index(fi.Name(), ".") != 0 {
				if err != nil {
					return errors.New("failed to get image size: " + err.Error())
				}

				name := strings.Split(fi.Name(), ".tar.gz")[0]

				size := fi.Size()

				created := int(time.Since(fi.ModTime()).Hours() / 24)
				var timeUnit string
				if created > 1 {
					timeUnit = strconv.Itoa(created) + " days ago"
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
							return errors.New("failed to generate image hash: " + err.Error())
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
						if err != nil {
							return errors.New(err.Error())
						}
					} else {
						return errors.New("couldn't load image hash: " + err.Error())
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
		return errors.New("failed to connect to host: " + err.Error())
	}

	if short {
		fmt.Println(info.IPv4)
		return nil
	}

	if info.State == "Stopped" {
		return errors.New("cannot connect to Bravetools remote, ensure it is up and running")
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
		return errors.New("cannot connect to Bravetools remote, ensure it is up and running")
	}

	units, err := GetUnits(bh.Remote)
	if err != nil {
		return errors.New("Failed to list units: " + err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Status", "IPv4", "Volumes", "Ports"})
	for _, u := range units {
		name := u.Name
		status := u.Status
		address := u.Address

		disk := ""
		for _, diskDevice := range u.Disk {
			if diskDevice.Name != "" {
				disk += diskDevice.Name + ":" + diskDevice.Source + "->" + diskDevice.Path + "\n"
			}
		}

		proxy := ""
		for _, proxyDevice := range u.Proxy {
			if proxyDevice.Name != "" {
				connectIP := strings.Split(proxyDevice.ConnectIP, ":")[2]
				listenIP := strings.Split(proxyDevice.ListenIP, ":")[2]
				proxy += connectIP + ":" + listenIP + "\n"
			}
		}

		r := []string{name, status, address, disk, proxy}
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

// UmountShare ..
func (bh *BraveHost) UmountShare(unit string, target string) error {
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
			return errors.New("could not check directory: " + err.Error())
		}
		output = strings.Trim(output, "\n")

		hostOs := runtime.GOOS
		if hostOs == "windows" {
			path = strings.Replace(path, string(filepath.Separator), "/", -1)
		}

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
			return errors.New("failed to umount " + target + ": " + err.Error())
		}
	}

	volume, _ := GetVolume(bh.Settings.StoragePool.Name, bh.Remote)
	if len(volume.UsedBy) == 0 {
		DeleteVolume(bh.Settings.StoragePool.Name, volume, bh.Remote)

		return nil
	}

	return nil
}

// MountShare ..
func (bh *BraveHost) MountShare(source string, destUnit string, destPath string) error {

	names, err := GetUnits(bh.Remote)
	if err != nil {
		return errors.New("faild to access units")
	}

	var found = false
	for _, n := range names {
		if n.Name == destUnit {
			found = true
			break
		}
	}
	if !found {
		return errors.New("unit not found")
	}

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
	sharedDirectory = filepath.Join("/home/ubuntu", "volumes", sharedDirectory)

	switch backend {
	case "multipass":

		hostOs := runtime.GOOS
		if hostOs == "windows" {
			sourcePath = filepath.FromSlash(sourcePath)
			destPath = strings.Replace(destPath, string(filepath.Separator), "/", -1)
			sharedDirectory = strings.Replace(sharedDirectory, string(filepath.Separator), "/", -1)
		}

		if sourceUnit == "" {
			err := shared.ExecCommand("multipass",
				"mount",
				sourcePath,
				bh.Settings.Name+":"+sharedDirectory)
			if err != nil {
				return errors.New("Failed to initialize mount on host :" + err.Error())
			}

			err = MountDirectory(sharedDirectory, destUnit, destPath, bh.Remote)
			if err != nil {
				err = shared.ExecCommand("multipass",
					"umount",
					bh.Settings.Name+":"+sharedDirectory)
				if err != nil {
					return err
				}
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

	unitList, err := GetUnits(bh.Remote)
	if err != nil {
		return errors.New("failed to list existing units: " + err.Error())
	}

	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	if !shared.StringInSlice(name, unitNames) {
		return errors.New("unit " + name + " does not exist")
	}

	err = DeleteUnit(name, bh.Remote)
	if err != nil {
		return errors.New("failed to delete unit: " + err.Error())
	}

	// Deleting unit from databse

	userHome, err := os.UserHomeDir()
	if err != nil {
		return errors.New("failed to get home directory")
	}
	dbPath := path.Join(userHome, shared.BraveDB)
	database, err := db.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database %s", dbPath)
	}

	err = db.DeleteUnitDB(database, name)
	if err != nil {
		return errors.New("failed to delete unit from database. Name: " + name + " Error: " + err.Error())
	}

	//fmt.Println("Unit deleted: ", name)
	return nil
}

// BuildImage creates an image based on Bravefile
func (bh *BraveHost) BuildImage(bravefile *shared.Bravefile) error {
	if strings.ContainsAny(bravefile.PlatformService.Name, "/_. !@£$%^&*(){}:;`~,?") {
		return errors.New("image names should not contain special characters")
	}

	if bravefile.PlatformService.Name == "" {
		return errors.New("service Name is empty")
	}

	err := checkUnits(bravefile.PlatformService.Name, bh)
	if err != nil {
		return err
	}

	originalHostImageList, err := listHostImages(bh.Remote)
	if err != nil {
		log.Fatal(err)
	}

	processInterruptHandler(originalHostImageList, bravefile, bh)

	switch bravefile.Base.Location {
	case "public":
		_, err = importLXD(bravefile, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return err
		}

		err = Start(bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return err
		}
	case "github":
		_, err = importGitHub(bravefile, bh)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return err
		}

		err = Start(bravefile.PlatformService.Name, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return err
		}
	case "local":
		_, err = importLocal(bravefile, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return err
		}
	default:
		return fmt.Errorf("base image location %q not supported", bravefile.Base.Location)
	}

	pMan := bravefile.SystemPackages.Manager

	switch pMan {
	case "":
		// No package manager - if packages are to be installed, raise error
		if len(bravefile.SystemPackages.System) > 0 {
			return errors.New("package manager not specified - cannot install packages")
		}
	case "apk":
		_, err := Exec(bravefile.PlatformService.Name, []string{"apk", "update", "--no-cache"}, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return errors.New("failed to update repositories: " + err.Error())
		}

		args := []string{"apk", "--no-cache", "add"}
		args = append(args, bravefile.SystemPackages.System...)

		if len(args) > 3 {
			status, err := Exec(bravefile.PlatformService.Name, args, bh.Remote)

			if err != nil {
				cleanupBuild(originalHostImageList, bravefile, bh)
				return errors.New("failed to install packages: " + err.Error())
			}
			if status > 0 {
				cleanupBuild(originalHostImageList, bravefile, bh)
				return errors.New(shared.Fatal("failed to install packages"))
			}
		}

	case "apt":
		_, err := Exec(bravefile.PlatformService.Name, []string{"apt", "update"}, bh.Remote)
		if err != nil {
			cleanupBuild(originalHostImageList, bravefile, bh)
			return errors.New("failed to update repositories: " + err.Error())
		}

		args := []string{"apt", "install"}
		args = append(args, bravefile.SystemPackages.System...)

		if len(args) > 2 {
			args = append(args, "--yes")
			status, err := Exec(bravefile.PlatformService.Name, args, bh.Remote)

			if err != nil {
				cleanupBuild(originalHostImageList, bravefile, bh)
				return errors.New("failed to install packages: " + err.Error())
			}
			if status > 0 {
				cleanupBuild(originalHostImageList, bravefile, bh)
				return errors.New(shared.Fatal("failed to install packages"))
			}
		}
	default:
		return fmt.Errorf("package manager %q not recognized", pMan)
	}

	// Go through "Copy" section
	err = bravefileCopy(bravefile.Copy, bravefile.PlatformService.Name, bh.Remote)
	if err != nil {
		cleanupBuild(originalHostImageList, bravefile, bh)
		return err
	}

	// Go through "Run" section
	status, err := bravefileRun(bravefile.Run, bravefile.PlatformService.Name, bh.Remote)
	if err != nil {
		cleanupBuild(originalHostImageList, bravefile, bh)
		return errors.New("failed to execute command: " + err.Error())
	}
	if status > 0 {
		cleanupBuild(originalHostImageList, bravefile, bh)
		return errors.New(shared.Fatal("non-zero exit code: " + strconv.Itoa(status)))
	}

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	unitFingerprint, err := Publish(bravefile.PlatformService.Name, bravefile.PlatformService.Version, bh.Remote)
	if err != nil {
		cleanupBuild(originalHostImageList, bravefile, bh)
		return errors.New("failed to publish image: " + err.Error())
	}

	err = ExportImage(unitFingerprint, bravefile.PlatformService.Name+"-"+bravefile.PlatformService.Version, bh.Remote)
	if err != nil {
		cleanupBuild(originalHostImageList, bravefile, bh)
		return errors.New("failed to export image: " + err.Error())
	}

	cleanupBuild(originalHostImageList, bravefile, bh)

	home, _ := os.UserHomeDir()
	localImageFile := bravefile.PlatformService.Name + "-" + bravefile.PlatformService.Version + ".tar.gz"
	localHashFile := localImageFile + ".md5"

	imageHash, err := shared.FileHash(localImageFile)
	if err != nil {
		return errors.New("failed to generate image hash: " + err.Error())
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
		return errors.New("failed to copy image archive to local storage: " + err.Error())
	}

	err = shared.CopyFile(localHashFile, home+shared.ImageStore+localHashFile)
	if err != nil {
		return errors.New("failed to copy images hash into local storage: " + err.Error())
	}

	err = os.Remove(localImageFile)
	if err != nil {
		return errors.New("failed to clean up image archive: " + err.Error())
	}

	err = os.Remove(localHashFile)
	if err != nil {
		return errors.New("failed to clean up image hash: " + err.Error())
	}

	return nil
}

// PublishUnit publishes unit to image
func (bh *BraveHost) PublishUnit(name string, backend Backend) error {
	_, err := backend.Info()
	if err != nil {
		return errors.New("failed to get host info: " + err.Error())
	}

	timestamp := time.Now()

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	fmt.Println("Publishing unit ...")

	var unitFingerprint string
	unitFingerprint, err = Publish(name, timestamp.Format("20060102150405"), bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(unitFingerprint, bh.Remote)
		return errors.New("failed to publish image: " + err.Error())
	}

	fmt.Println("Exporting archive ...")
	err = ExportImage(unitFingerprint, name+"-"+timestamp.Format("20060102150405"), bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(unitFingerprint, bh.Remote)
		return errors.New("failed to export unit: " + err.Error())
	}

	fmt.Println("Cleaning ...")
	DeleteImageByFingerprint(unitFingerprint, bh.Remote)

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
	var fingerprint string

	homeDir, _ := os.UserHomeDir()
	if unitParams.PlatformService.Image == "" {
		return errors.New("unit image name cannot be empty")
	}
	image := homeDir + shared.ImageStore + unitParams.PlatformService.Image + ".tar.gz"

	// Resource checks
	err := CheckResources(image, backend, unitParams, bh)
	if err != nil {
		return err
	}

	fingerprint, err = ImportImage(image, unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		return errors.New("failed to import image: " + err.Error())
	}

	err = LaunchFromImage(unitParams.PlatformService.Name, unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to launch unit: " + err.Error())
	}

	err = AttachNetwork(unitParams.PlatformService.Name, bh.Settings.Name+"br0", "eth0", "eth0", bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to attach network: " + err.Error())
	}

	err = ConfigDevice(unitParams.PlatformService.Name, "eth0", unitParams.PlatformService.IP, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to set IP: " + err.Error())
	}

	err = Stop(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to stop unit: " + err.Error())
	}

	err = Start(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("Failed to restart unit: " + err.Error())
	}

	user, err := user.Current()
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return err
	}

	var uid string
	var gid string

	hostOs := runtime.GOOS
	if hostOs == "windows" {
		uidParts := strings.Split(user.Uid, "-")
		gidParts := strings.Split(user.Gid, "-")

		uid = uidParts[len(uidParts)-1]
		gid = gidParts[len(gidParts)-1]
	} else {
		uid = user.Uid
		gid = user.Gid
	}

	vm := *NewLxd(bh.Settings)
	_, whichLxc, err := lxdCheck(vm)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return err
	}

	clientVersion, _, err := checkLXDVersion(whichLxc)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return err
	}

	// uid and gid mapping is not allowed in non-snap LXD. Shares can be created, but they are read-only in a unit.
	var config map[string]string
	if clientVersion <= 303 {
		config = map[string]string{
			"limits.cpu":       unitParams.PlatformService.Resources.CPU,
			"limits.memory":    unitParams.PlatformService.Resources.RAM,
			"security.nesting": "false",
			"nvidia.runtime":   "false",
		}
	} else {
		config = map[string]string{
			"limits.cpu":       unitParams.PlatformService.Resources.CPU,
			"limits.memory":    unitParams.PlatformService.Resources.RAM,
			"raw.idmap":        "both " + uid + " " + gid,
			"security.nesting": "false",
			"nvidia.runtime":   "false",
		}
	}

	if unitParams.PlatformService.Docker == "yes" {
		config["security.nesting"] = "true"
	}

	if unitParams.PlatformService.Resources.GPU == "yes" {
		config["nvidia.runtime"] = "true"
		device := map[string]string{"type": "gpu"}
		err = AddDevice(unitParams.PlatformService.Name, "gpu", device, bh.Remote)
		if err != nil {
			DeleteImageByFingerprint(fingerprint, bh.Remote)
			return errors.New("failed to add GPU device: " + err.Error())
		}
	}

	err = SetConfig(unitParams.PlatformService.Name, config, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("error configuring unit: " + err.Error())
	}

	err = Stop(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to stop unit: " + err.Error())
	}

	err = Start(unitParams.PlatformService.Name, bh.Remote)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to restart unit: " + err.Error())
	}

	ports := unitParams.PlatformService.Ports
	if len(ports) > 0 {
		for _, p := range ports {
			ps := strings.Split(p, ":")
			if len(ps) < 2 {
				DeleteImageByFingerprint(fingerprint, bh.Remote)
				DeleteUnit(unitParams.PlatformService.Name, bh.Remote)
				return errors.New("invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
			}

			err := addIPRules(unitParams.PlatformService.Name, ps[1], ps[0], bh)
			if err != nil {
				DeleteImageByFingerprint(fingerprint, bh.Remote)
				delErr := DeleteUnit(unitParams.PlatformService.Name, bh.Remote)
				if delErr != nil {
					DeleteImageByFingerprint(fingerprint, bh.Remote)
					return errors.New("failed to delete unit: " + delErr.Error())
				}
				return errors.New("unable to add Proxy Device: " + err.Error())
			}
		}
	}

	// Add unit into database

	var braveUnit db.BraveUnit
	userHome, err := os.UserHomeDir()
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return errors.New("failed to get home directory")
	}
	dbPath := path.Join(userHome, shared.BraveDB)

	database, err := db.OpenDB(dbPath)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		return fmt.Errorf("failed to open database %s", dbPath)
	}

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
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		DeleteUnit(unitParams.PlatformService.Name, bh.Remote)
		return errors.New("failed to serialize unit data")
	}
	braveUnit.Data = data

	_, err = db.InsertUnitDB(database, braveUnit)
	if err != nil {
		DeleteImageByFingerprint(fingerprint, bh.Remote)
		DeleteUnit(unitParams.PlatformService.Name, bh.Remote)
		return errors.New("failed to insert unit to database: " + err.Error())
	}

	DeleteImageByFingerprint(fingerprint, bh.Remote)

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
