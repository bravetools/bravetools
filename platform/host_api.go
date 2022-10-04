package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"path/filepath"

	"github.com/bravetools/bravetools/db"
	"github.com/bravetools/bravetools/shared"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
)

// Functions exposed to commands.go

// DeleteImageName deletes image by name
func (bh *BraveHost) DeleteImageByName(name string) error {
	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	images, _ := listHostImages(lxdServer)
	if len(images) > 0 {
		err := DeleteImageByName(lxdServer, name)
		if err != nil {
			return errors.New("image: " + err.Error())
		}
	}

	return nil
}

// DeleteImage delete image by fingerprint
func (bh *BraveHost) DeleteImageByFingerprint(fingerprint string) error {
	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	err = DeleteImageByFingerprint(lxdServer, fingerprint)
	if err != nil {
		return errors.New("failed to delete image: " + err.Error())
	}

	return nil
}

// DeleteHostImages removes all LXC images from host
func (bh *BraveHost) DeleteHostImages() error {
	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	err = deleteHostImages(lxdServer)
	if err != nil {
		return err
	}
	return nil
}

// AddRemote sets connection to Brave platform
func (bh *BraveHost) AddRemote() error {
	err := AddRemote(bh.Remote, bh.Settings.Trust)
	if err != nil {
		return errors.New("failed to add remote host: " + err.Error())
	}

	return nil
}

// ImportLocalImage import tarball into local images folder
func (bh *BraveHost) ImportLocalImage(sourcePath string) error {
	home, _ := os.UserHomeDir()
	imageStore := path.Join(home, shared.ImageStore)

	_, imageName := filepath.Split(sourcePath)

	image, err := ImageFromFilename(imageName)
	if err != nil {
		return err
	}

	if imageExists(image) {
		return errors.New("image " + imageName + " already exists in local image store")
	}

	imagePath := filepath.Join(imageStore, image.ToBasename())
	if err != nil {
		return err
	}
	hashFile := imagePath + ".md5"

	err = shared.CopyFile(sourcePath, imagePath)
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
	imageStore := path.Join(home, shared.ImageStore)

	// We're only interested in images and not MD5 checksums
	images, err := shared.WalkMatch(imageStore, "*.tar.gz")
	if err != nil {
		return errors.New("failed to access images folder: " + err.Error())
	}

	if len(images) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Image", "Version", "Arch", "Created", "Size", "Hash"})

		for _, i := range images {
			image, err := ImageFromFilename(filepath.Base(i))
			if err != nil {
				return fmt.Errorf("failed to parse image filename schema from %q", i)
			}

			fi, err := os.Stat(i)
			if strings.Index(fi.Name(), ".") != 0 {
				if err != nil {
					return errors.New("failed to get image size: " + err.Error())
				}

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

				hashString, err := getImageHash(image)
				if err != nil {
					return err
				}

				r := []string{image.Name, image.Version, image.Architecture, timeUnit, shared.FormatByteCountSI(size), hashString}
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
	image, err := ParseImageString(name)
	if err != nil {
		return err
	}
	if !imageExists(image) {
		image, err = ParseLegacyImageString(name)
		if err != nil {
			return err
		}
		if !imageExists(image) {
			return fmt.Errorf("image %q does not exist", name)
		}
	}
	imagePath, err := getImageFilepath(image)
	if err != nil {
		return err
	}
	imageHash := imagePath + ".md5"

	err = os.Remove(imagePath)
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
func (bh *BraveHost) HostInfo(short bool) error {
	info, err := bh.Backend.Info()
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
		info.Disk.UsedStorage + " of " + info.Disk.TotalStorage,
		info.Memory.UsedStorage + " of " + info.Memory.TotalStorage, info.CPU}

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
func (bh *BraveHost) ListUnits(backend Backend, remoteName string) error {
	info, err := backend.Info()
	if err != nil {
		return err
	}

	if info.State == "Stopped" {
		return errors.New("cannot connect to Bravetools remote, ensure it is up and running")
	}

	var units []shared.BraveUnit

	if remoteName != "" {
		deployRemote, err := LoadRemoteSettings(remoteName)
		if err != nil {
			return err
		}

		lxdServer, err := GetLXDInstanceServer(deployRemote)
		if err != nil {
			return err
		}

		deployProfile := deployRemote.Profile
		if deployProfile == "" {
			deployProfile = deployRemote.Profile
		}

		units, err = GetUnits(lxdServer, deployProfile)
		if err != nil {
			return errors.New("Failed to list units: " + err.Error())
		}
	} else {
		// Load all units on all remotes

		remoteNames, err := ListRemotes()
		if err != nil {
			return err
		}

		for i := range remoteNames {
			deployRemote, err := LoadRemoteSettings(remoteNames[i])
			if err != nil {
				return err
			}

			// If no auth, this isn't a deploy remote unless unix protocol
			if (deployRemote.key == "" || deployRemote.cert == "") && deployRemote.Protocol != "unix" {
				continue
			}

			lxdServer, err := GetLXDInstanceServer(deployRemote)
			if err != nil {
				log.Printf("failed to connect to %q remote, skipping", deployRemote.Name)
				continue
			}

			remoteUnits, err := GetUnits(lxdServer, deployRemote.Profile)
			if err != nil {
				return errors.New("Failed to list units: " + err.Error())
			}

			// Prefix unit name with remote name
			if deployRemote.Name != shared.BravetoolsRemote {
				for j := range remoteUnits {
					remoteUnits[j].Name = deployRemote.Name + ":" + remoteUnits[j].Name
				}
			}
			units = append(units, remoteUnits...)
		}
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

	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	switch backend {
	case "multipass":
		path, err := DeleteDevice(lxdServer, unit, target)
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
		_, err := DeleteDevice(lxdServer, unit, target)
		if err != nil {
			return errors.New("failed to umount " + target + ": " + err.Error())
		}
	}

	volume, _ := GetVolume(lxdServer, bh.Settings.StoragePool.Name)
	if len(volume.UsedBy) == 0 {
		DeleteVolume(lxdServer, bh.Settings.StoragePool.Name, volume)

		return nil
	}

	return nil
}

// MountShare ..
func (bh *BraveHost) MountShare(source string, destUnit string, destPath string) error {

	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	names, err := GetUnits(lxdServer, bh.Remote.Profile)
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

			err = MountDirectory(lxdServer, sharedDirectory, destUnit, destPath)
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
			err := createSharedVolume(lxdServer,
				bh.Settings.StoragePool.Name,
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
			err := MountDirectory(lxdServer, sourcePath, destUnit, destPath)
			if err != nil {
				return errors.New("Failed to mount " + source + " to " + destUnit + ":" + destPath + " : " + err.Error())
			}
		} else {
			err := createSharedVolume(lxdServer,
				bh.Settings.StoragePool.Name,
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

	remoteName, name := ParseRemoteName(name)

	// If local remote, ensure the VM is started
	if remoteName == shared.BravetoolsRemote {
		info, err := bh.Backend.Info()
		if err != nil {
			return errors.New("failed to get host info: " + err.Error())
		}
		if strings.ToLower(info.State) == "stopped" {
			return errors.New("Backend is stopped")
		}
	}

	remote, err := LoadRemoteSettings(remoteName)
	if err != nil {
		return err
	}

	lxdServer, err := GetLXDInstanceServer(remote)
	if err != nil {
		return err
	}

	// Remote profile to filter units by - try using deploy profile. If not, fallback to brave host profile.
	remoteProfile := remote.Profile
	if remoteProfile == "" {
		remoteProfile = bh.Remote.Profile
	}

	unitList, err := GetUnits(lxdServer, remoteProfile)
	if err != nil {
		return errors.New("failed to list existing units: " + err.Error())
	}

	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	if !shared.StringInSlice(name, unitNames) {
		return errors.New("unit " + name + " does not exist")
	}

	err = DeleteUnit(lxdServer, name)
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

	return nil
}

type ImageExistsError struct {
	Name string
}

func (e *ImageExistsError) Error() string {
	return fmt.Sprintf("image %q already exists", e.Name)
}

// BuildImage creates an image based on Bravefile
func (bh *BraveHost) BuildImage(bravefile shared.Bravefile) error {

	var imageStruct BravetoolsImage
	var err error

	// The image to build - if not in build section, use Image defined in Service section
	imageString := bravefile.Image
	if imageString == "" {
		imageString = bravefile.PlatformService.Image
	}

	err = bravefile.ValidateBuild()
	if err != nil {
		return fmt.Errorf("failed to build image: %s", err)
	}

	// If version explicitly provided separately this is a legacy Bravefile
	if bravefile.PlatformService.Version == "" || bravefile.Image != "" {
		imageStruct, err = ParseImageString(imageString)
	} else {
		imageStruct, err = ParseLegacyImageString(imageString)
	}
	if err != nil {
		return err
	}

	// Image architecture must match base image architecture
	baseImageStruct, err := ParseImageString(bravefile.Base.Image)
	if err != nil {
		return err
	}
	if imageStruct.Architecture != baseImageStruct.Architecture {
		return fmt.Errorf("target image architecture [%s] does not match base image [%s]", imageStruct.Architecture, baseImageStruct.Architecture)
	}

	if imageExists(imageStruct) {
		return &ImageExistsError{Name: imageStruct.String()}
	}

	fmt.Println(shared.Info("Building Image: " + imageStruct.String()))

	bravefile.PlatformService.Name = "brave-build-" + strings.ReplaceAll(strings.ReplaceAll(imageStruct.ToBasename(), "_", "-"), ".", "-")

	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	err = checkUnits(lxdServer, bravefile.PlatformService.Name, bh.Remote.Profile)
	if err != nil {
		return err
	}

	// Intercept SIGINT, propagate cancel and cleanup artefacts
	var imageFingerprint string

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			fmt.Println("Interrupting build and cleaning artefacts")
			cancel()
		}
	}()

	// Setup build cleanup code
	defer func() {
		DeleteUnit(lxdServer, bravefile.PlatformService.Name)
		DeleteImageByFingerprint(lxdServer, imageFingerprint)
	}()

	// If base image location not provided, attempt to infer it
	if bravefile.Base.Location == "" {
		bravefile.Base.Location, err = resolveBaseImageLocation(bravefile.Base.Image)
		if err != nil {
			return fmt.Errorf("base image %q does not exist: %s", bravefile.Base.Image, err.Error())
		}
	}

	switch bravefile.Base.Location {
	case "public":
		// Check disk space
		publicLxd, err := GetSimplestreamsLXDSever("https://images.linuxcontainers.org", nil)
		if err != nil {
			return err
		}

		img, err := GetImageByAlias(publicLxd, bravefile.Base.Image)
		if err != nil {
			return err
		}
		err = CheckStoragePoolSpace(lxdServer, bh.Settings.StoragePool.Name, img.Size)
		if err != nil {
			return err
		}

		imageFingerprint, err = importLXD(ctx, lxdServer, &bravefile, bh.Remote.Profile)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		err = Start(lxdServer, bravefile.PlatformService.Name)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}
	case "github":
		imageFingerprint, err = importGitHub(ctx, lxdServer, &bravefile, bh, bh.Remote.Profile, bh.Remote.Storage)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		err = Start(lxdServer, bravefile.PlatformService.Name)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}
	case "local":
		// Check disk space
		localBaseImage, err := ParseImageString(bravefile.Base.Image)
		if err != nil {
			return err
		}
		if !imageExists(localBaseImage) {
			// Check legacy bravefile
			localBaseImage, err = ParseLegacyImageString(bravefile.Base.Image)
			if err != nil {
				return err
			}
			if !imageExists(localBaseImage) {
				return fmt.Errorf("base image %q required for building image %q does not exist", localBaseImage.String(), imageStruct.String())
			}
		}

		imgSize, err := localImageSize(localBaseImage)
		if err != nil {
			return err
		}
		err = CheckStoragePoolSpace(lxdServer, bh.Settings.StoragePool.Name, imgSize)
		if err != nil {
			return err
		}

		imageFingerprint, err = importLocal(ctx, lxdServer, &bravefile, bh.Remote.Profile, bh.Remote.Storage)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
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
		_, err := Exec(ctx, lxdServer, bravefile.PlatformService.Name, []string{"apk", "update", "--no-cache"}, ExecArgs{})
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return errors.New("failed to update repositories: " + err.Error())
		}

		args := []string{"apk", "--no-cache", "add"}
		args = append(args, bravefile.SystemPackages.System...)

		if len(args) > 3 {
			status, err := Exec(ctx, lxdServer, bravefile.PlatformService.Name, args, ExecArgs{})

			if err := shared.CollectErrors(err, ctx.Err()); err != nil {
				return errors.New("failed to install packages: " + err.Error())
			}
			if status > 0 {
				return errors.New(shared.Fatal("failed to install packages"))
			}
		}

	case "apt":
		_, err := Exec(ctx, lxdServer, bravefile.PlatformService.Name, []string{"apt", "update"}, ExecArgs{})
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return errors.New("failed to update repositories: " + err.Error())
		}

		args := []string{"apt", "install"}
		args = append(args, bravefile.SystemPackages.System...)

		if len(args) > 2 {
			args = append(args, "--yes")
			status, err := Exec(ctx, lxdServer, bravefile.PlatformService.Name, args, ExecArgs{})

			if err := shared.CollectErrors(err, ctx.Err()); err != nil {
				return errors.New("failed to install packages: " + err.Error())
			}
			if status > 0 {
				return errors.New(shared.Fatal("failed to install packages"))
			}
		}
	default:
		return fmt.Errorf("package manager %q not recognized", pMan)
	}

	// Go through "Copy" section
	err = bravefileCopy(ctx, lxdServer, bravefile.Copy, bravefile.PlatformService.Name)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return err
	}

	// Go through "Run" section
	err = bravefileRun(ctx, lxdServer, bravefile.Run, bravefile.PlatformService.Name)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New(shared.Fatal("failed to execute command: " + err.Error()))
	}

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	unitFingerprint, err := Publish(lxdServer, bravefile.PlatformService.Name, imageStruct.String())
	defer DeleteImageByFingerprint(lxdServer, unitFingerprint)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to publish image: " + err.Error())
	}

	err = ExportImage(lxdServer, unitFingerprint, imageStruct.ToBasename())
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to export image: " + err.Error())
	}

	home, _ := os.UserHomeDir()
	localImageFile := imageStruct.ToBasename() + ".tar.gz"
	localHashFile := localImageFile + ".md5"

	defer func() {
		if err := os.Remove(localImageFile); err != nil {
			fmt.Println("failed to clean up image archive: " + err.Error())
		}
	}()

	imageHash, err := shared.FileHash(localImageFile)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to generate image hash: " + err.Error())
	}

	fmt.Println(imageHash)

	// Write image hash to a file
	f, err := os.Create(localHashFile)
	if err != nil {
		return errors.New(err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("failed to close image hash file: " + err.Error())
		}
		if err := os.Remove(localHashFile); err != nil {
			fmt.Println("failed to clean up image hash: " + err.Error())
		}
	}()

	_, err = f.WriteString(imageHash)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New(err.Error())
	}

	err = shared.CopyFile(localImageFile, path.Join(home, shared.ImageStore, localImageFile))
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to copy image archive to local storage: " + err.Error())
	}

	err = shared.CopyFile(localHashFile, path.Join(home, shared.ImageStore, localHashFile))
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to copy images hash into local storage: " + err.Error())
	}

	return nil
}

// PublishUnit publishes unit to image
func (bh *BraveHost) PublishUnit(name string, backend Backend) error {
	_, err := backend.Info()
	if err != nil {
		return errors.New("failed to get host info: " + err.Error())
	}

	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}

	timestamp := time.Now()

	// Create an image based on running container and export it. Image saved as tar.gz in project local directory.
	fmt.Printf("Publishing unit %q\n", name)

	unitFingerprint, err := Publish(lxdServer, name, timestamp.Format("20060102150405"))
	defer DeleteImageByFingerprint(lxdServer, unitFingerprint)
	if err != nil {
		return errors.New("failed to publish image: " + err.Error())
	}

	fmt.Println("Exporting archive ...")
	err = ExportImage(lxdServer, unitFingerprint, name+"-"+timestamp.Format("20060102150405"))
	if err != nil {
		return errors.New("failed to export unit: " + err.Error())
	}

	fmt.Println("Cleaning ...")

	return nil
}

// StopUnit stops unit using name
func (bh *BraveHost) StopUnit(name string) error {

	remoteName, name := ParseRemoteName(name)

	// If local remote, ensure the VM is started
	if remoteName == shared.BravetoolsRemote {
		info, err := bh.Backend.Info()
		if err != nil {
			return errors.New("Failed to get host info: " + err.Error())
		}
		if strings.ToLower(info.State) == "stopped" {
			return errors.New("Backend is stopped")
		}
	}

	remote, err := LoadRemoteSettings(remoteName)
	if err != nil {
		return err
	}

	lxdServer, err := GetLXDInstanceServer(remote)
	if err != nil {
		return err
	}

	fmt.Println("Stopping unit: ", name)
	err = Stop(lxdServer, name)
	if err != nil {
		return errors.New("Failed to stop unit: " + err.Error())
	}

	return nil
}

// StartUnit restarts unit if running and starts if stopped.
func (bh *BraveHost) StartUnit(name string) error {
	remoteName, name := ParseRemoteName(name)

	// If local remote, ensure the VM is started
	if remoteName == shared.BravetoolsRemote {
		info, err := bh.Backend.Info()
		if err != nil {
			return errors.New("Failed to get host info: " + err.Error())
		}
		if strings.ToLower(info.State) == "stopped" {
			return errors.New("Backend is stopped")
		}
	}

	remote, err := LoadRemoteSettings(remoteName)
	if err != nil {
		return err
	}

	lxdServer, err := GetLXDInstanceServer(remote)
	if err != nil {
		return err
	}

	fmt.Println("Starting unit: ", name)
	err = Start(lxdServer, name)
	if err != nil {
		return errors.New("Failed to start unit: " + err.Error())
	}

	return nil
}

// InitUnit starts unit from supplied image
func (bh *BraveHost) InitUnit(backend Backend, unitParams shared.Service) (err error) {
	// Check for missing mandatory fields
	err = unitParams.ValidateDeploy()
	if err != nil {
		return err
	}

	fmt.Println(shared.Info("Deploying Unit " + unitParams.Name))

	var imageStruct BravetoolsImage

	// If version explicitly provided separately this is a legacy Bravefile
	if unitParams.Version == "" {
		imageStruct, err = ParseImageString(unitParams.Image)
	} else {
		imageStruct, err = ParseLegacyImageString(unitParams.Image)
	}
	if err != nil {
		return err
	}

	if !imageExists(imageStruct) {
		return fmt.Errorf("image %q does not exist", imageStruct.String())
	}

	// Connect to deploy target remote
	deployRemoteName, unitName := ParseRemoteName(unitParams.Name)
	unitParams.Name = unitName

	deployRemote, err := LoadRemoteSettings(deployRemoteName)
	if err != nil {
		return fmt.Errorf("failed to load remote %q for requested unit %q: %s", deployRemoteName, unitName, err.Error())
	}

	// Load remote defaults for LXD resources for deployment (profile, network, storage) if not specified in Bravefile unitParams
	if unitParams.Profile == "" {
		unitParams.Profile = deployRemote.Profile
	}
	if unitParams.Network == "" {
		unitParams.Network = deployRemote.Network
	}
	if unitParams.Storage == "" {
		unitParams.Storage = deployRemote.Storage
	}

	// As last resort if not provided in Bravefile or remote, try the Brave host settings - mostly for backward compatability
	if unitParams.Profile == "" && unitParams.Network == "" && unitParams.Storage == "" {
		unitParams.Profile = bh.Settings.Profile
		unitParams.Network = bh.Settings.Name
		unitParams.Storage = bh.Settings.StoragePool.Name
	}

	lxdServer, err := GetLXDInstanceServer(deployRemote)
	if err != nil {
		return err
	}

	// Check if a unit with this name already exists - we don't want to delete it
	err = checkUnits(lxdServer, unitName, unitParams.Profile)
	if err != nil {
		return err
	}

	image, err := getImageFilepath(imageStruct)
	if err != nil {
		return err
	}
	fingerprint, err := shared.FileSha256Hash(image)
	if err != nil {
		return fmt.Errorf("failed to obtain image hash %q", unitParams.Image)
	}
	defer DeleteImageByFingerprint(lxdServer, fingerprint)

	// Resource checks
	err = CheckMemory(lxdServer, unitParams.Resources.RAM)
	if err != nil {
		return err
	}
	err = CheckHostPorts(deployRemote.URL, unitParams.Ports)
	if err != nil {
		return err
	}

	// Intercept SIGINT and cancel context, triggering cleanup of resources
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			fmt.Println("Interrupting deployment and cleaning artefacts")
			cancel()
		}
	}()

	_, err = ImportImage(lxdServer, image, unitName)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to import image: " + err.Error())
	}

	// Launch unit and set up cleanup code to delete it if an error encountered during deployment
	err = LaunchFromImage(lxdServer, unitName, unitName, unitParams.Profile, unitParams.Storage)
	defer func() {
		if err != nil {
			delErr := DeleteUnit(lxdServer, unitName)
			if delErr != nil {
				fmt.Println("failed to delete unit: " + delErr.Error())
			}
		}
	}()
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to launch unit: " + err.Error())
	}

	err = AttachNetwork(lxdServer, unitName, unitParams.Network, "eth0", "eth0")
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to attach network: " + err.Error())
	}

	// Assign static IP
	err = ConfigDevice(lxdServer, unitName, "eth0", unitParams.IP)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to set IP: " + err.Error())
	}

	err = Stop(lxdServer, unitName)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to stop unit: " + err.Error())
	}

	err = Start(lxdServer, unitName)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("Failed to restart unit: " + err.Error())
	}

	user, err := user.Current()
	if err != nil {
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

	serverVersion, err := GetLXDServerVersion(lxdServer)
	if err != nil {
		return errors.New("failed to get server version: " + err.Error())
	}

	// uid and gid mapping is not allowed in non-snap LXD. Shares can be created, but they are read-only in a unit.
	var config map[string]string
	if serverVersion <= 303 {
		config = map[string]string{
			"limits.cpu":       unitParams.Resources.CPU,
			"limits.memory":    unitParams.Resources.RAM,
			"security.nesting": "false",
			"nvidia.runtime":   "false",
		}
	} else {
		config = map[string]string{
			"limits.cpu":       unitParams.Resources.CPU,
			"limits.memory":    unitParams.Resources.RAM,
			"raw.idmap":        "both " + uid + " " + gid,
			"security.nesting": "false",
			"nvidia.runtime":   "false",
		}
	}

	if unitParams.Docker == "yes" {
		config["security.nesting"] = "true"
	}

	if unitParams.Resources.GPU == "yes" {
		config["nvidia.runtime"] = "true"
		device := map[string]string{"type": "gpu"}
		err = AddDevice(lxdServer, unitName, "gpu", device)
		if err != nil {
			return errors.New("failed to add GPU device: " + err.Error())
		}
	}

	err = SetConfig(lxdServer, unitName, config)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("error configuring unit: " + err.Error())
	}

	err = Stop(lxdServer, unitName)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to stop unit: " + err.Error())
	}

	err = Start(lxdServer, unitName)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to restart unit: " + err.Error())
	}

	ports := unitParams.Ports
	if len(ports) > 0 {
		for _, p := range ports {
			ps := strings.Split(p, ":")
			if len(ps) < 2 {
				err = errors.New("invalid port forwarding definition. Appropriate format is UNIT_PORT:HOST_PORT")
				return
			}

			err = addIPRules(lxdServer, unitName, ps[1], ps[0])
			if err = shared.CollectErrors(err, ctx.Err()); err != nil {
				return errors.New("unable to add Proxy Device: " + err.Error())
			}
		}
	}

	err = postdeploy(ctx, lxdServer, &unitParams)
	if err = shared.CollectErrors(err, ctx.Err()); err != nil {
		return err
	}

	// Add unit into database

	var braveUnit db.BraveUnit
	userHome, err := os.UserHomeDir()
	if err != nil {
		return errors.New("failed to get home directory")
	}
	dbPath := path.Join(userHome, shared.BraveDB)

	database, err := db.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database %s", dbPath)
	}

	uuid, _ := uuid.NewUUID()
	braveUnit.UID = uuid.String()
	braveUnit.Name = unitName
	braveUnit.Date = time.Now().String()

	var unitData db.UnitData
	unitData.CPU, _ = strconv.Atoi(unitParams.Resources.CPU)
	unitData.RAM = unitParams.Resources.RAM
	unitData.IP = unitParams.IP
	unitData.Image = unitParams.Image

	data, err := json.Marshal(unitData)
	if err != nil {
		return errors.New("failed to serialize unit data")
	}
	braveUnit.Data = data

	_, err = db.InsertUnitDB(database, braveUnit)
	if err != nil {
		return errors.New("failed to insert unit to database: " + err.Error())
	}

	return nil
}

func (bh *BraveHost) Compose(backend Backend, composeFile *shared.ComposeFile) (err error) {

	// Compose runs from parent directory of compose file
	workingDir, err := filepath.Abs(filepath.Dir(composeFile.Path))
	if err != nil {
		return err
	}
	startDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(workingDir)
	defer os.Chdir(startDir)

	// Order services by deps
	topologicalOrdering, err := composeFile.TopologicalOrdering()
	if err != nil {
		return err
	}

	// Remove base-only services if all images depending on them already exist
	for _, baseService := range getBaseOnlyServices(composeFile) {
		dependentServices, err := getBuildDependents(baseService, composeFile)
		if err != nil {
			return err
		}
		if len(dependentServices) == 0 {
			serviceIdx, err := shared.StrSliceIndexOf(topologicalOrdering, baseService)
			if err != nil {
				return err
			}
			topologicalOrdering = append(topologicalOrdering[:serviceIdx], topologicalOrdering[serviceIdx+1:]...)
		}
	}

	// Validate Services
	for _, serviceName := range topologicalOrdering {
		service, exist := composeFile.Services[serviceName]
		if !exist {
			err = fmt.Errorf("service name %q does not exist in Services", serviceName)
			return err
		}
		err = service.ValidateDeploy()
		if err != nil {
			err = fmt.Errorf("failed to deploy service %q: %s", serviceName, err)
			return err
		}
	}

	// (Optionally build) and deploy each service
	for _, serviceName := range topologicalOrdering {
		service := composeFile.Services[serviceName]

		// Load bravefile settings as defaults, overwrite if specified in composefile
		if service.Bravefile != "" {
			if service.Build || service.Base {
				err = service.BravefileBuild.ValidateBuild()
				if err != nil {
					return fmt.Errorf("invalid Bravefile for service %q: %s", service.Name, err)
				}

				// Switch to build context dir
				buildDir := service.Context
				if buildDir == "" {
					buildDir, err = filepath.Abs(filepath.Dir(service.Bravefile))
				}
				os.Chdir(buildDir)

				err = bh.BuildImage(*service.BravefileBuild)
				switch errType := err.(type) {
				case nil:
					// Cleanup image later if error in compose
					defer func() {
						if err != nil {
							bh.DeleteLocalImage(service.Image)
						}
					}()
				case *ImageExistsError:
					// If image already exists continue and log the skip
					err = nil
					fmt.Printf("image %q already exists - skipping build\n", errType.Name)
				default:
					// Stop on unknown err
					return err
				}

				os.Chdir(workingDir)

				if service.Base && !service.Build {
					defer func() {
						bh.DeleteLocalImage(service.Image)
					}()
				}
			}
		}

		// Only deploy service if it isn't a base image used during build only
		if !service.Base {
			// Deploy context - use Context if provided, else Bravefile if present, else current dir
			deployDir := service.Context
			if deployDir == "" {
				if service.Bravefile != "" {
					deployDir, err = filepath.Abs(filepath.Dir(service.Bravefile))
					if err != nil {
						return err
					}
				} else {
					deployDir = "."
				}
			}
			os.Chdir(deployDir)

			// Cleanup each unit if error in compose
			err = bh.InitUnit(backend, service.Service)
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					bh.DeleteUnit(service.Name)
				}
			}()

			os.Chdir(workingDir)
		}

	}
	return nil
}
