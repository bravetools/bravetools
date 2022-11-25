package platform

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bravetools/bravetools/shared"
)

const defaultImageVersion = "untagged"

type BravetoolsImage struct {
	Name         string
	Version      string
	Architecture string
}

func ParseImageString(imageString string) (imageStruct BravetoolsImage, err error) {
	// Remove remote if present
	_, imageString = ParseRemoteName(imageString)

	// Image schema: image/version/arch
	split := strings.SplitN(imageString, "/", 3)

	if split[0] == "" {
		return imageStruct, fmt.Errorf("image name not provided in %q - bravetools image schema is remote:<image_name>[/version/arch]", imageString)
	}

	// Default struct
	imageStruct = BravetoolsImage{
		Name:         split[0],
		Version:      "",
		Architecture: "",
	}

	// Override defaults if provided
	if len(split) >= 3 {
		imageStruct.Architecture = split[2]
	}

	if len(split) >= 2 {
		imageStruct.Version = split[1]
	}

	if err := validateImage(imageStruct); err != nil {
		return imageStruct, fmt.Errorf("image %q is not a valid string: %s", imageString, err)
	}

	return imageStruct, nil
}

func ParseLegacyImageString(imageString string) (imageStruct BravetoolsImage, err error) {
	// Remove remote if present
	_, imageString = ParseRemoteName(imageString)

	// Legacy Bravefile - these have the version prepended to end of name and no arch
	split := strings.Split(imageString, "-")
	if split[0] == "" {
		return imageStruct, errors.New("image name not provided")
	}
	if len(split) == 1 {
		return imageStruct, fmt.Errorf("failed to parse legacy Bravefile image field %q - expected %q at end", imageString, "-[version]")
	}

	// Default struct
	imageStruct = BravetoolsImage{
		Name:         strings.Join(split[:len(split)-1], "-"),
		Version:      split[len(split)-1],
		Architecture: "",
	}

	if err := validateImage(imageStruct); err != nil {
		return imageStruct, fmt.Errorf("image %q is not a valid string: %s", imageString, err)
	}

	return imageStruct, nil
}
func validateImage(imageStruct BravetoolsImage) error {
	// Cannot have empty name
	if imageStruct.Name == "" {
		return errors.New("image cannot have an empty name")
	}

	// Check Name, Version and Architecture fields for non-allowed characters
	for _, char := range imageStruct.Name {
		if !validImageFieldChar(char) {
			return fmt.Errorf("character %q is not valid in image name field", char)
		}
	}
	for _, char := range imageStruct.Version {
		if !validImageFieldChar(char) {
			return fmt.Errorf("character %q is not valid in image version field", char)
		}
	}
	for _, char := range imageStruct.Architecture {
		// Underscore allowed for arch - special case
		if !validImageFieldChar(char) && char != '_' {
			return fmt.Errorf("character %q is not valid in image architecture field", char)
		}
	}

	return nil
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
	fields := []string{}
	for _, field := range []string{imageStruct.Name, imageStruct.Version, imageStruct.Architecture} {
		if field == "" {
			break
		}
		fields = append(fields, field)
	}
	return strings.Join(fields, "/")
}

func ImageFromFilename(filename string) (BravetoolsImage, error) {
	filename = strings.TrimSuffix(filename, ".tar.gz")
	split := strings.SplitN(filename, "_", 3)

	image := BravetoolsImage{
		Name:         split[0],
		Version:      defaultImageVersion,
		Architecture: "",
	}

	if len(split) > 1 {
		image.Version = split[1]
	}
	if len(split) > 2 {
		image.Architecture = split[2]
	}

	if err := validateImage(image); err != nil {
		return image, fmt.Errorf("image %q is not a valid string: %s", image, err)
	}

	return image, nil
}

func ImageFromLegacyFilename(filename string) (BravetoolsImage, error) {
	// Legacy filenames are not delimited by underscores
	// Final "-" is followed by version - no arch
	filename = strings.TrimSuffix(filename, ".tar.gz")
	split := strings.Split(filename, "-")

	image := BravetoolsImage{
		Name:         split[0],
		Version:      defaultImageVersion,
		Architecture: "",
	}

	if len(split) > 1 {
		image.Name = strings.Join(split[:len(split)-1], "-")
		image.Version = split[len(split)-1]
	}

	if err := validateImage(image); err != nil {
		return image, fmt.Errorf("image %q is not a valid string: %s", image, err)
	}

	return image, nil
}

type multipleImageMatches struct {
	image         BravetoolsImage
	matchingPaths []string
}

func (e multipleImageMatches) Error() string {
	// Multiple matches - ambiguous result. Return formatted error with the options.
	imageStrings := make([]string, len(e.matchingPaths))
	for _, path := range e.matchingPaths {
		img, err := ImageFromFilename(filepath.Base(path))
		if err != nil {
			imageStrings = append(imageStrings, path)
		} else {
			imageStrings = append(imageStrings, img.String())
		}
	}

	return fmt.Sprintf("multiple matches for image %q in image store - specify version and/or architecture.\nMatches:%s", e.image, strings.Join(imageStrings, "\n"))
}

// matchLocalImagePath attempts to find candidates for the provided image definition using regex matching.
// If more than one candidate file exists a formatted error of type 'multipleImageMatches' is returned.
func matchLocalImagePath(image BravetoolsImage) (string, error) {

	// Before querying candidates using regex, attempt to exactly match the provided image definition
	if path, err := localImagePath(image); err == nil {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to access bravetools image store: %s", err)
	}

	var fileRegexArr []string
	for _, field := range []string{image.Name, image.Version, image.Architecture} {
		if field == "" {
			fileRegexArr = append(fileRegexArr, "*")
		} else {
			fileRegexArr = append(fileRegexArr, field)
		}
	}

	imagePath := filepath.Join(homeDir, shared.ImageStore, strings.Join(fileRegexArr, "_")) + ".tar.gz"

	matches, err := filepath.Glob(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to search bravetools image store: %s", err)
	}

	nmatches := len(matches)
	switch {
	case nmatches == 1:
		return matches[0], nil
	case nmatches > 1:
		// Multiple matches - ambiguous result. Return formatted error with the options.
		err = multipleImageMatches{image: image, matchingPaths: matches}
		return "", err
	}

	return "", fmt.Errorf("failed to retrieve path for image %s, version: %s, arch: %s ", image.Name, image.Version, image.Architecture)
}

// localImagePath gets the exact image filepath matching the definition if it exists - no regex matching is performed
func localImagePath(image BravetoolsImage) (string, error) {
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
	return "", fmt.Errorf("failed to retrieve path for image %s, version: %s, arch: %s ", image.Name, image.Version, image.Architecture)
}

// hashImage calculates the md5 hash of the provided BravetoolsImage and stores it in a file.
// If a file with a hash for this image already exists the hash will not be recalculated.
func hashImage(image BravetoolsImage) (string, error) {
	localImageFile, err := matchLocalImagePath(image)
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
	imagePath, err := matchLocalImagePath(image)
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

func resolveBaseImageLocation(imageString string, architecture string) (location string, err error) {

	remote, imageString := ParseRemoteName(imageString)

	if remote == "github.com" {
		return "github", nil
	}

	imageStruct, err := ParseImageString(imageString)
	if err != nil {
		return "", err
	}

	if _, err = matchLocalImagePath(imageStruct); err == nil {
		return "local", nil
	}

	remoteList, err := ListRemotes()
	if err != nil {
		return "", err
	}
	for _, remoteName := range remoteList {
		if remote == remoteName && remote != shared.BravetoolsRemote {
			return "private", nil
		}
	}

	// Check for legacy image field
	imageStruct, err = ParseLegacyImageString(imageString)
	if err == nil {
		if _, err = matchLocalImagePath(imageStruct); err == nil {
			return "local", nil
		}
	}

	// Query public remote for alias
	publicLxd, err := GetSimplestreamsLXDSever("https://images.linuxcontainers.org", nil)
	if err != nil {
		return "", err
	}
	if _, err := GetFingerprintByAlias(publicLxd, imageString, architecture); err == nil {
		return "public", nil
	}

	return "", fmt.Errorf("image %q location could not be resolved", imageString)
}

// GetBravefileFromLXD generates a Bravefile for import of images from an LXD server
func GetBravefileFromLXD(name string) (*shared.Bravefile, error) {
	imageRemoteName, name := ParseRemoteName(name)

	image, err := ParseImageString(name)
	if err != nil {
		return nil, err
	}

	baseImageName := image.String()
	baseLocation := "public"
	if imageRemoteName != shared.BravetoolsRemote {
		baseImageName = imageRemoteName + ":" + baseImageName
		baseLocation = "private"
	}

	bravefile := shared.NewBravefile()

	bravefile.Image = image.String()
	bravefile.Base.Image = baseImageName
	bravefile.Base.Location = baseLocation

	bravefile.PlatformService.Name = ""
	bravefile.PlatformService.Image = image.String()

	return bravefile, nil
}
