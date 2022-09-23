package platform

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/bravetools/bravetools/shared"
)

const defaultImageVersion = "1.0"

type BravetoolsImage struct {
	Name         string
	Version      string
	Architecture string
}

func ParseImageString(imageString string) (imageStruct BravetoolsImage, err error) {
	// Remove remote if present
	_, imageString = ParseRemoteName(imageString)

	// Image schema: image/version/arch
	split := strings.SplitN(imageString, "/", 4)

	if len(split) > 4 {
		return imageStruct, fmt.Errorf("unrecongized format - bravetools image schema is remote:<image_name>[/version/arch]")
	}

	if split[0] == "" {
		return imageStruct, fmt.Errorf("image name not provided in %q - bravetools image schema is remote:<image_name>[/version/arch]", imageString)
	}

	// Default struct
	// Architecture defaults to runtime arch
	// Version defaults to 1.0
	imageStruct = BravetoolsImage{
		Name:         split[0],
		Version:      defaultImageVersion,
		Architecture: runtime.GOARCH,
	}

	// Override defaults if provided
	if len(split) >= 3 {
		imageStruct.Architecture = split[2]
	}

	if len(split) >= 2 {
		imageStruct.Version = split[1]
	}

	return imageStruct, nil
}

func (imageStruct BravetoolsImage) ToBasename() string {
	fields := []string{imageStruct.Name, imageStruct.Version, imageStruct.Architecture}
	return strings.Join(fields, "_")
}

func (imageStruct BravetoolsImage) String() string {
	fields := []string{imageStruct.Name, imageStruct.Version, imageStruct.Architecture}
	return strings.Join(fields, "/")
}

func ImageFromFilename(filename string) (BravetoolsImage, error) {
	filename = strings.TrimSuffix(filename, ".tar.gz")
	split := strings.Split(filename, "_")
	if len(split) > 3 {
		return BravetoolsImage{}, fmt.Errorf("filename %q does not conform to image format '<image_name>_<version>_<arch>", filename)
	}

	image := BravetoolsImage{
		Name:         split[0],
		Version:      defaultImageVersion,
		Architecture: runtime.GOARCH,
	}

	if len(split) > 1 {
		image.Version = split[1]
	}
	if len(split) > 2 {
		image.Architecture = split[2]
	}

	return image, nil
}

func imageExists(image BravetoolsImage) bool {
	homeDir, _ := os.UserHomeDir()
	imagePath := path.Join(homeDir, shared.ImageStore, image.ToBasename()+".tar.gz")
	return shared.FileExists(imagePath)
}

func localImageSize(image BravetoolsImage) (bytes int64, err error) {
	homeDir, _ := os.UserHomeDir()
	imagePath := path.Join(homeDir, shared.ImageStore, image.ToBasename()+".tar.gz")

	info, err := os.Stat(imagePath)
	if err != nil {
		return -1, err
	}
	if info.IsDir() {
		return -1, fmt.Errorf("expected image path %q to be a file, found dir", imagePath)
	}

	return info.Size(), nil
}

func resolveBaseImageLocation(imageString string) (location string, err error) {

	remote, imageString := ParseRemoteName(imageString)

	if remote == "github.com" {
		return "github", nil
	}

	imageStruct, err := ParseImageString(imageString)
	if err != nil {
		return "", err
	}

	if imageExists(imageStruct) {
		return "local", nil
	}

	// Query public remote for alias
	publicLxd, err := GetSimplestreamsLXDSever("https://images.linuxcontainers.org", nil)
	if err != nil {
		return "", err
	}
	if _, err := GetFingerprintByAlias(publicLxd, imageString); err == nil {
		return "public", nil
	}

	return "", fmt.Errorf("image %q location could not be resolved", imageString)
}
