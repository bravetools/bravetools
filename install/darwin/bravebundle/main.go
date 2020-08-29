package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	iconFilePath string
)

func init() {
	flag.StringVar(&iconFilePath, "icon", "", "The application icon")
}

func main() {
	flag.Parse()
	if iconFilePath == "" {
		log.Println("ERROR: icon file required")
		flag.PrintDefaults()
		return
	}
	makeAppIcons(iconFilePath)
}

func makeAppIcons(icoFile string) error {
	var err error
	iconFileName := filepath.Base(iconFilePath)
	fmt.Println("Icon file: ", iconFileName)
	resFolder := filepath.Join("macos", "Bravetools.app", "Contents", "Resources")

	tmpFolder := filepath.Join(resFolder, "tmp")
	err = os.MkdirAll(tmpFolder, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpFolder)

	iconset := filepath.Join(tmpFolder, "icon.iconset")
	err = os.Mkdir(iconset, 0755)
	if err != nil {
		return err
	}
	sizes := []int{16, 32, 64, 128, 256, 512, 1024}
	for i, size := range sizes {
		nameSize := size
		var suffix string
		if i > 0 {
			nameSize = sizes[i-1]
			suffix = "@2x"
		}

		iconName := fmt.Sprintf("icon_%dx%d%s.png", nameSize, nameSize, suffix)
		outIconFile := filepath.Join(iconset, iconName)

		sizeStr := fmt.Sprintf("%d", size)
		cmd := exec.Command("sips", "-z", sizeStr, sizeStr, iconFilePath, "--out", outIconFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("running sips: %v", err)
		}

		if i > 0 && i < len(sizes)-1 {
			stdName := fmt.Sprintf("icon_%dx%d.png", size, size)
			err := copyFile(outIconFile, filepath.Join(iconset, stdName), nil)
			if err != nil {
				return fmt.Errorf("copying icon file: %v", err)
			}
		}
	}

	icnsFile := filepath.Join(resFolder, "icon.icns")
	cmd := exec.Command("iconutil", "-c", "icns", "-o", icnsFile, iconset)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("running iconutil: %v", err)
	}

	return nil
}

func copyFile(from, to string, fromInfo os.FileInfo) error {
	log.Printf("[INFO] Copying %s to %s", from, to)

	if fromInfo == nil {
		var err error
		fromInfo, err = os.Stat(from)
		if err != nil {
			return err
		}
	}

	// open source file
	fsrc, err := os.Open(from)
	if err != nil {
		return err
	}

	// create destination file, with identical permissions
	fdest, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fromInfo.Mode()&os.ModePerm)
	if err != nil {
		fsrc.Close()
		if _, err2 := os.Stat(to); err2 == nil {
			return fmt.Errorf("opening destination (which already exists): %v", err)
		}
		return err
	}

	// copy the file and ensure it gets flushed to disk
	if _, err = io.Copy(fdest, fsrc); err != nil {
		fsrc.Close()
		fdest.Close()
		return err
	}
	if err = fdest.Sync(); err != nil {
		fsrc.Close()
		fdest.Close()
		return err
	}

	// close both files
	if err = fsrc.Close(); err != nil {
		fdest.Close()
		return err
	}
	if err = fdest.Close(); err != nil {
		return err
	}

	return nil
}
