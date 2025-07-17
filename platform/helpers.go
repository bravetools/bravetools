package platform

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/bravetools/bravetools/shared"
	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

// Private Helpers

func getCurrentUsername() (username string, err error) {

	user, err := user.Current()
	if err != nil {
		return "", err
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}

	username = reg.ReplaceAllString(user.Username, "")

	return "bravetools-" + username, nil
}

// createSharedVolume creates a volume in storage pool and mounts it to both source unit and target unit
func createSharedVolume(lxdServer lxd.InstanceServer,
	storagePoolName string,
	sourceUnit string,
	sourcePath string,
	destUnit string,
	destPath string) error {

	sourcePath = cleanMountTargetPath(sourcePath)
	destPath = cleanMountTargetPath(destPath)

	volumeName := getDiskDeviceHash(sourceUnit, sourcePath)

	newVolume := api.StorageVolumesPost{
		Name:        volumeName,
		Type:        "custom",
		ContentType: "filesystem",
	}
	err := lxdServer.CreateStoragePoolVolume(storagePoolName, newVolume)
	if err != nil {
		return err
	}

	sourceShareSettings := map[string]string{
		"path":   sourcePath,
		"pool":   storagePoolName,
		"source": volumeName,
		"type":   "disk",
	}

	// 2. Add storage volume as a disk device to source unit
	sourceDeviceName := getDiskDeviceHash(sourceUnit, sourcePath)
	err = AddDevice(lxdServer, sourceUnit, sourceDeviceName, sourceShareSettings)
	if err != nil {
		if err := lxdServer.DeleteStoragePoolVolume(storagePoolName, "custom", volumeName); err != nil {
			log.Printf("failed to cleanup storage volume %q from pool %q", volumeName, storagePoolName)
		}
		return err
	}

	destShareSettings := map[string]string{
		"path":   destPath,
		"pool":   storagePoolName,
		"source": volumeName,
		"type":   "disk",
	}

	// 3. Add storage volume as a disk device to target unit
	destDeviceName := getDiskDeviceHash(destUnit, destPath)
	err = AddDevice(lxdServer, destUnit, destDeviceName, destShareSettings)
	if err != nil {
		return errors.New("failed to mount to destination: " + err.Error())
	}

	return nil
}

func needTransferImage(bravefile shared.Bravefile) bool {
	// The image to build - if not in build section, use Image defined in Service section
	imageString := bravefile.Image
	if imageString == "" {
		imageString = bravefile.PlatformService.Image
	}

	destRemoteName, _ := ParseRemoteName(imageString)

	// If no remote store specified for image nothing to do
	return destRemoteName != shared.BravetoolsRemote
}

func buildImage(bh *BraveHost, bravefile *shared.Bravefile) error {

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
	if !bravefile.IsLegacy() {
		imageStruct, err = ParseImageString(imageString)
	} else {
		imageStruct, err = ParseLegacyImageString(imageString)
	}
	if err != nil {
		return err
	}

	// Use bravetools host LXD instance to build
	lxdServer, err := GetLXDInstanceServer(bh.Remote)
	if err != nil {
		return err
	}
	// Images must match the build server arch
	buildServerArch, err := GetLXDServerArch(lxdServer)
	if err != nil {
		return err
	}

	// Set output image architecture based on server arch if not provided and set default version if missing
	if imageStruct.Architecture == "" {
		imageStruct.Architecture = buildServerArch
	}
	if imageStruct.Version == "" {
		imageStruct.Version = defaultImageVersion
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

	// If image already exists in local store, check for remote dest - if exists, push image there, else error
	if _, err := localImagePath(imageStruct); err == nil {
		return &ImageExistsError{Name: imageStruct.String()}
	}

	fmt.Println(shared.Info("Building Image: " + imageStruct.String()))

	bravefile.PlatformService.Name = "brave-build-" + strings.ReplaceAll(strings.ReplaceAll(imageStruct.ToBasename(), "_", "-"), ".", "-")

	err = checkUnits(lxdServer, bravefile.PlatformService.Name, bh.Remote.Profile)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return err
	}

	// Setup build cleanup code
	defer func() {
		DeleteUnit(lxdServer, bravefile.PlatformService.Name)
		DeleteImageByFingerprint(lxdServer, imageFingerprint)
	}()

	// If base image location not provided, attempt to infer it
	if bravefile.Base.Location == "" {
		bravefile.Base.Location, err = resolveBaseImageLocation(bravefile.Base.Image, buildServerArch, bh.Settings.PublicImageRemote)
		if err != nil {
			return fmt.Errorf("base image %q does not exist: %s", bravefile.Base.Image, err.Error())
		}
	}

	switch bravefile.Base.Location {
	case "public", "private":
		var sourceImageServer lxd.ImageServer

		// Connect to image source LXD server
		if bravefile.Base.Location == "public" {
			sourceImageServer, err = GetSimplestreamsLXDSever(bh.Settings.PublicImageRemote, nil)
			if err != nil {
				return err
			}
		}
		if bravefile.Base.Location == "private" {
			var imageRemoteName string
			imageRemoteName, bravefile.Base.Image = ParseRemoteName(bravefile.Base.Image)

			imageRemote, err := LoadRemoteSettings(imageRemoteName)
			if err != nil {
				return err
			}

			// Connect to remote server - authenticate if not public
			if imageRemote.Public {
				sourceImageServer, err = GetLXDImageSever(imageRemote)
			} else {
				sourceImageServer, err = GetLXDInstanceServer(imageRemote)
			}
			if err != nil {
				return err
			}
		}

		// Check disk space
		img, err := GetImageByAlias(sourceImageServer, bravefile.Base.Image, buildServerArch)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		err = CheckStoragePoolSpace(lxdServer, bh.Remote.Storage, img.Size)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		imageFingerprint, err = LaunchFromImage(lxdServer, sourceImageServer, bravefile.Base.Image, bravefile.PlatformService.Name, bh.Remote.Profile, bh.Remote.Storage)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		err = Start(lxdServer, bravefile.PlatformService.Name)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}
	case "github":
		imageFingerprint, err = importGitHub(ctx, lxdServer, bravefile, bh, bh.Remote.Profile, bh.Remote.Storage)
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
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}
		if localBaseImage.Architecture == "" {
			localBaseImage.Architecture = buildServerArch
		}
		if _, err = matchLocalImagePath(localBaseImage); err != nil {
			// In case of multiple possible matches ask user to specify rather than proceed to legacy image parsing
			if errors.As(err, &multipleImageMatches{}) {
				return err
			}

			// Check legacy bravefile
			var parseErr error
			localBaseImage, parseErr = ParseLegacyImageString(bravefile.Base.Image)
			if parseErr == nil {
				if _, legacyErr := matchLocalImagePath(localBaseImage); legacyErr != nil {
					return legacyErr
				}
			} else {
				return err
			}
		}

		imgSize, err := localImageSize(localBaseImage)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}
		err = CheckStoragePoolSpace(lxdServer, bh.Remote.Storage, imgSize)
		if err := shared.CollectErrors(err, ctx.Err()); err != nil {
			return err
		}

		imageFingerprint, err = importLocal(ctx, lxdServer, bravefile, bh.Remote.Profile, bh.Remote.Storage)
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
	unitFingerprint, err := Publish(lxdServer, bravefile.PlatformService.Name, imageStruct.ToBasename())
	defer DeleteImageByFingerprint(lxdServer, unitFingerprint)
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to publish image: " + err.Error())
	}

	err = ExportImage(lxdServer, unitFingerprint, imageStruct.ToBasename())
	if err := shared.CollectErrors(err, ctx.Err()); err != nil {
		return errors.New("failed to export image: " + err.Error())
	}

	err = importImageFile(ctx, imageStruct)
	if err != nil {
		return errors.New("failed to copy image file to bravetools image store: " + err.Error())
	}

	return nil
}

func TransferImage(sourceRemote Remote, bravefile shared.Bravefile) error {
	var imageStruct BravetoolsImage
	var err error

	// The image to build - if not in build section, use Image defined in Service section
	imageString := bravefile.Image
	if imageString == "" {
		imageString = bravefile.PlatformService.Image
	}

	// If version explicitly provided separately this is a legacy Bravefile
	if !bravefile.IsLegacy() {
		imageStruct, err = ParseImageString(imageString)
	} else {
		imageStruct, err = ParseLegacyImageString(imageString)
	}
	if err != nil {
		return err
	}

	destRemoteName, _ := ParseRemoteName(imageString)

	// If no remote store specified for image nothing to do
	if destRemoteName == shared.BravetoolsRemote {
		return nil
	}

	// Use bravetools host LXD instance to build
	lxdServer, err := GetLXDInstanceServer(sourceRemote)
	if err != nil {
		return err
	}

	// Set output image architecture based on server arch if not provided and set default version if missing
	if imageStruct.Architecture == "" {
		buildServerArch, err := GetLXDServerArch(lxdServer)
		if err != nil {
			return err
		}
		imageStruct.Architecture = buildServerArch
	}
	if imageStruct.Version == "" {
		imageStruct.Version = defaultImageVersion
	}

	imgPath, err := localImagePath(imageStruct)
	if err != nil {
		return err
	}

	fmt.Println(shared.Info(fmt.Sprintf("Pushing image to remote %q", destRemoteName)))

	destRemote, err := LoadRemoteSettings(destRemoteName)
	if err != nil {
		return err
	}

	destServer, err := GetLXDInstanceServer(destRemote)
	if err != nil {
		return err
	}

	// Import image to local LXD server to transfer to remote
	imageFingerprint, err := ImportImage(lxdServer, imgPath, imageStruct.String())
	if err != nil {
		return err
	}

	// If the remote to push image to is not the same as bravehost remote, cleanup and push
	if sourceRemote.Name != destRemoteName {
		defer func() {
			DeleteImageByFingerprint(lxdServer, imageFingerprint)
		}()

		err = CopyImage(lxdServer, destServer, imageFingerprint, imageStruct.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func importGitHub(ctx context.Context, lxdServer lxd.InstanceServer, bravefile *shared.Bravefile, bh *BraveHost, profileName string, storagePool string) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}

	path := bravefile.Base.Image
	if !strings.HasPrefix(path, "github.com/") {
		path = "github.com/" + path
	}
	remoteBravefile, err := shared.GetBravefileFromGitHub(path)
	if err != nil {
		return fingerprint, err
	}

	var imageStruct BravetoolsImage

	// If version explicitly provided separately this is a legacy Bravefile
	if !remoteBravefile.PlatformService.IsLegacy() {
		imageStruct, err = ParseImageString(remoteBravefile.PlatformService.Image)
	} else {
		imageStruct, err = ParseLegacyImageString(remoteBravefile.PlatformService.Image)
	}
	if err != nil {
		return fingerprint, err
	}

	if _, err = matchLocalImagePath(imageStruct); err != nil {
		err = bh.BuildImage(*remoteBravefile)
		if err != nil {
			return fingerprint, err
		}
	} else {
		fmt.Println("Found local image " + imageStruct.String() + ". Skipping GitHub build")
	}

	remoteBravefile.Base.Image = imageStruct.String()
	remoteBravefile.PlatformService.Name = bravefile.PlatformService.Name
	// Since we are using new image format above, we need to set version to "" to prevent parsing as legacy image name
	remoteBravefile.PlatformService.Version = ""

	fingerprint, err = importLocal(ctx, lxdServer, remoteBravefile, profileName, storagePool)
	return fingerprint, err
}

func importLocal(ctx context.Context, lxdServer lxd.InstanceServer, bravefile *shared.Bravefile, profileName string, storagePool string) (fingerprint string, err error) {
	if err = ctx.Err(); err != nil {
		return "", err
	}
	var imageStruct BravetoolsImage

	imageStruct, err = ParseImageString(bravefile.Base.Image)
	if err != nil {
		return "", err
	}
	if imageStruct.Architecture == "" {
		imageStruct.Architecture, err = GetLXDServerArch(lxdServer)
		if err != nil {
			return "", fmt.Errorf("failed to get lxd build server arch: %s", err)
		}
	}

	path, err := matchLocalImagePath(imageStruct)
	if err != nil {
		if errors.As(err, &multipleImageMatches{}) {
			return "", err
		}

		var legacyParseErr error
		var legacyMatchErr error
		imageStruct, legacyParseErr = ParseLegacyImageString(bravefile.Base.Image)
		if legacyParseErr != nil {
			return "", err
		}

		path, legacyMatchErr = matchLocalImagePath(imageStruct)
		if legacyMatchErr != nil {
			return "", err
		}
	}

	fingerprint, err = shared.FileSha256Hash(path)
	if err != nil {
		return fingerprint, err
	}

	_, err = ImportImage(lxdServer, path, bravefile.Base.Image)
	if err != nil {
		return fingerprint, errors.New("failed to import image: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	_, err = LaunchFromImage(lxdServer, lxdServer, bravefile.Base.Image, bravefile.PlatformService.Name, profileName, storagePool)
	if err != nil {
		return fingerprint, errors.New("failed to launch unit: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	err = Start(lxdServer, bravefile.PlatformService.Name)
	if err != nil {
		return fingerprint, errors.New("failed to start a unit: " + err.Error())
	}

	if err = ctx.Err(); err != nil {
		return fingerprint, err
	}

	return fingerprint, nil
}

// postdeploy copy files and run commands on running service
func postdeploy(ctx context.Context, lxdServer lxd.InstanceServer, unitConfig *shared.Service) (err error) {

	if unitConfig.Postdeploy.Copy != nil {
		err = bravefileCopy(ctx, lxdServer, unitConfig.Postdeploy.Copy, unitConfig.Name)
		if err != nil {
			return err
		}
	}

	if unitConfig.Postdeploy.Run != nil {
		err = bravefileRun(ctx, lxdServer, unitConfig.Postdeploy.Run, unitConfig.Name)
		if err != nil {
			return errors.New(shared.Fatal("failed to execute command: " + err.Error()))
		}
	}

	return nil
}

func bravefileCopy(ctx context.Context, lxdServer lxd.InstanceServer, copy []shared.CopyCommand, service string) error {
	dir, _ := os.Getwd()
	for _, c := range copy {
		if err := ctx.Err(); err != nil {
			return err
		}

		source := c.Source
		sourcePath := filepath.Join(dir, source)

		target := c.Target
		_, err := Exec(ctx, lxdServer, service, []string{"mkdir", "-p", target}, ExecArgs{})
		if err != nil {
			return errors.New("Failed to create target directory: " + err.Error())
		}

		fi, err := os.Lstat(sourcePath)
		if err != nil {
			return errors.New("Failed to read file " + sourcePath + ": " + err.Error())
		}

		if fi.IsDir() {
			err = Push(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push directory: " + err.Error())
			}
		} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			err = SymlinkPush(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push symlink: " + err.Error())
			}
		} else {
			err = FilePush(lxdServer, service, sourcePath, target)
			if err != nil {
				return errors.New("Failed to push file: " + err.Error())
			}
		}

		if c.Action != "" {
			_, err = Exec(ctx, lxdServer, service, []string{"sh", "-c", c.Action}, ExecArgs{})
			if err != nil {
				return errors.New("Failed to execute action: " + err.Error())
			}
		}
	}

	return nil
}

func bravefileRun(ctx context.Context, lxdServer lxd.InstanceServer, run []shared.RunCommand, service string) (err error) {
	for _, c := range run {
		if err = ctx.Err(); err != nil {
			return err
		}

		var command string
		var content string

		if c.Command != "" {
			command = c.Command
		}

		args := []string{command}
		if len(c.Args) > 0 {
			args = append(args, c.Args...)
		}
		if c.Content != "" {
			content = c.Content
			args = append(args, content)
		}

		status, err := Exec(ctx, lxdServer, service, args, ExecArgs{env: c.Env, detach: c.Detach})
		if err != nil {
			return err
		}
		if status > 0 {
			return fmt.Errorf("non-zero exit code %d for command %q", status, strings.Join(args, " "))
		}
	}

	return err
}

func cleanUnusedStoragePool(lxdServer lxd.InstanceServer, name string) {
	err := DeleteStoragePool(lxdServer, name)
	if err != nil {
		fmt.Println("Nothing to clean")
	}
}

// addIPRules adds firewall rule to the host iptable

func addIPRules(lxdServer lxd.InstanceServer, ct string, hostPort string, ctPort string) error {

	name := ct + "-proxy-" + hostPort + "-" + ctPort

	var config = make(map[string]string)

	config["type"] = "proxy"
	config["listen"] = "tcp:0.0.0.0:" + hostPort
	config["connect"] = "tcp:127.0.0.1:" + ctPort

	err := AddDevice(lxdServer, ct, name, config)
	if err != nil {
		return errors.New("failed to add proxy settings for unit " + err.Error())
	}

	return nil
}

func checkUnits(lxdServer lxd.InstanceServer, unitName string, profileName string) error {
	if unitName == "" {
		return errors.New("unit name cannot be empty")
	}

	// Unit Checks
	unitList, err := GetUnits(lxdServer, profileName)
	if err != nil {
		return err
	}

	var unitNames []string
	for _, u := range unitList {
		unitNames = append(unitNames, u.Name)
	}

	unitExists := shared.StringInSlice(unitName, unitNames)
	if unitExists {
		return errors.New("Unit " + unitName + " already exists on host")
	}

	return nil
}

func getBaseOnlyServices(composeFile *shared.ComposeFile) (serviceNames []string) {
	for serviceName := range composeFile.Services {
		if composeFile.Services[serviceName].Base && !composeFile.Services[serviceName].Build {
			serviceNames = append(serviceNames, serviceName)
		}
	}
	return serviceNames
}

func getBuildDependents(dependency string, composeFile *shared.ComposeFile) (serviceNames []string, err error) {
	for service := range composeFile.Services {
		var imageStruct BravetoolsImage

		// If version explicitly provided separately this is a legacy Bravefile
		if !composeFile.Services[service].IsLegacy() {
			imageStruct, err = ParseImageString(composeFile.Services[service].Image)
		} else {
			imageStruct, err = ParseLegacyImageString(composeFile.Services[service].Image)
		}
		if err != nil {
			return serviceNames, err
		}

		if _, err = matchLocalImagePath(imageStruct); err == nil {
			continue
		}
		for _, dependsOn := range composeFile.Services[service].Depends {
			if dependsOn == dependency {
				serviceNames = append(serviceNames, service)
			}
		}
	}
	return serviceNames, nil
}

func getDiskDeviceHash(unitName string, targetPath string) string {
	targetPath = cleanMountTargetPath(targetPath)
	return "brave_" + fmt.Sprintf("%x", sha256.Sum224([]byte(unitName+targetPath)))
}

func cleanMountTargetPath(targetPath string) string {
	targetPath = filepath.ToSlash(targetPath)
	targetPath = path.Clean(targetPath)
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}
	return targetPath
}

// importImageFile imports an LXD image file in the local directory into the bravetools image store
// The image file is cleaned up afterwards.
func importImageFile(ctx context.Context, imageStruct BravetoolsImage) error {
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
