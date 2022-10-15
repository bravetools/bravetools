package platform

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"unicode"

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

	if !validImageName(imageStruct) {
		return imageStruct, fmt.Errorf("image %q is not a valid string - fields must not contain special characters", imageString)
	}

	return imageStruct, nil
}

func ParseLegacyImageString(imageString string) (imageStruct BravetoolsImage, err error) {
	// Legacy Bravefile - these have the version prepended to end of name and no arch
	split := strings.Split(imageString, "-")
	if split[0] == "" {
		return imageStruct, errors.New("image name not provided")
	}
	if len(split) == 1 {
		return imageStruct, fmt.Errorf("failed to parse legacy Bravefile image field %q - expected %q at end", imageString, "-[version]")
	}

	// Default struct
	// Architecture defaults to runtime arch
	imageStruct = BravetoolsImage{
		Name:         strings.Join(split[:len(split)-1], "-"),
		Version:      split[len(split)-1],
		Architecture: runtime.GOARCH,
	}

	if !validImageName(imageStruct) {
		return imageStruct, fmt.Errorf("image %q is not a valid string - fields must not contain special characters", imageString)
	}

	return imageStruct, nil
}
func validImageName(imageStruct BravetoolsImage) bool {
	// Check Name, Version and Architecture fields for non-allowed characters
	for _, char := range imageStruct.Name {
		if !validImageFieldChar(char) {
			return false
		}
	}
	for _, char := range imageStruct.Version {
		if !validImageFieldChar(char) {
			return false
		}
	}
	for _, char := range imageStruct.Architecture {
		if !validImageFieldChar(char) {
			return false
		}
	}

	return true
}

func validImageFieldChar(char rune) bool {
	// Alpha-numeric is fine, along with '-' and '.'
	if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '-' && char != '.' {
		return false
	}
	return true
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

	// Legacy filenames are not delimited by underscores
	// Final "-" is followed by version - no arch
	if len(split) == 1 {
		split = strings.Split(filename, "-")
		image.Name = strings.Join(split[:len(split)-1], "-")
		image.Version = split[len(split)-1]
	}

	return image, nil
}

func imageExists(image BravetoolsImage) bool {
	_, err := getImageFilepath(image)
	return err == nil
}

func getImageFilepath(image BravetoolsImage) (string, error) {
	homeDir, _ := os.UserHomeDir()
	imagePath := path.Join(homeDir, shared.ImageStore, image.ToBasename()+".tar.gz")
	if shared.FileExists(imagePath) {
		return imagePath, nil
	}
	// Legacy filenames will not have arch
	imagePath = path.Join(homeDir, shared.ImageStore, image.Name+"-"+image.Version+".tar.gz")
	if shared.FileExists(imagePath) {
		return imagePath, nil
	}
	return "", fmt.Errorf("failed to retrieve path for image %s, version %s, arch %s ", image.Name, image.Version, image.Architecture)
}

func getImageHash(image BravetoolsImage) (string, error) {
	localImageFile, err := getImageFilepath(image)
	if err != nil {
		return "", err
	}
	hashFileName := localImageFile + ".md5"

	hash, err := ioutil.ReadFile(hashFileName)
	if err != nil {
		if os.IsNotExist(err) {

			imageHash, err := shared.FileHash(localImageFile)
			if err != nil {
				return "", errors.New("failed to generate image hash: " + err.Error())
			}

			f, err := os.Create(hashFileName)
			if err != nil {
				return "", errors.New(err.Error())
			}
			defer f.Close()

			_, err = f.WriteString(imageHash)
			if err != nil {
				return "", errors.New(err.Error())
			}

			hash, err = ioutil.ReadFile(hashFileName)
			if err != nil {
				return "", errors.New(err.Error())
			}
		} else {
			return "", errors.New("couldn't load image hash: " + err.Error())
		}
	}

	hashString := string(hash)
	hashString = strings.TrimRight(hashString, "\r\n")
	return hashString, nil
}

func localImageSize(image BravetoolsImage) (bytes int64, err error) {
	imagePath, err := getImageFilepath(image)
	if err != nil {
		return bytes, err
	}

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

	// Check for legacy image field
	imageStruct, err = ParseLegacyImageString(imageString)
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

// GetBravefileFromLXD generates a Bravefile for import of images from LXD repository
func GetBravefileFromLXD(name string) (*shared.Bravefile, error) {
	bravefile := shared.NewBravefile()

	image, err := ParseImageString(name)
	if err != nil {
		return nil, err
	}

	bravefile.Image = image.String()
	bravefile.Base.Image = image.String()
	bravefile.Base.Location = "public"

	bravefile.PlatformService.Name = ""
	bravefile.PlatformService.Image = image.String()

	return bravefile, nil
}
