package shared

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/briandowns/spinner"
	"gopkg.in/yaml.v2"
)

var (
	// Info ..
	Info = teal
	// Warn ..
	Warn = yellow
	// Fatal ..
	Fatal = red
)

var (
	black   = Color("\033[1;30m%s\033[0m")
	red     = Color("\033[1;31m%s\033[0m")
	green   = Color("\033[1;32m%s\033[0m")
	yellow  = Color("\033[1;33m%s\033[0m")
	purple  = Color("\033[1;34m%s\033[0m")
	magenta = Color("\033[1;35m%s\033[0m")
	teal    = Color("\033[1;36m%s\033[0m")
	white   = Color("\033[1;37m%s\033[0m")
)

// Color applies colors in terminal
func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func ping(ip string, port string) error {
	address, err := net.ResolveTCPAddr("tcp", ip+":"+port)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil
	}

	if conn != nil {
		defer conn.Close()
		return errors.New("Port " + port + " already assigned on host")
	}

	return err
}

// TCPPortStatus checks if multiple ports are available on the host
func TCPPortStatus(ip string, ports []string) error {
	for _, port := range ports {
		err := ping(ip, port)
		if err != nil {
			return err
		}

	}
	return nil
}

// GetBravefileFromGitHub reads bravefile from a github URL
func GetBravefileFromGitHub(name string) (*Bravefile, error) {
	var bravefile Bravefile
	var baseConfig string

	path := strings.SplitN(name, "/", -1)
	user := path[1]
	repository := path[2]
	project := strings.Join(path[3:], "/")

	url := "https://raw.githubusercontent.com/" + user + "/" + repository + "/master/" + project + "/Bravefile"

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		baseConfig = buf.String()
	}

	if len(baseConfig) == 0 {
		return nil, errors.New("Unable to download valid Bravefile. Check your URL")
	}

	err = yaml.Unmarshal([]byte(baseConfig), &bravefile)
	if err != nil {
		return nil, err
	}

	return &bravefile, nil
}

// GetBravefileFromLXD generates a Bravefile for import of images from LXD repository
func GetBravefileFromLXD(name string) (*Bravefile, error) {
	var bravefile Bravefile
	var baseConfig string

	dist := strings.SplitN(name, "/", -1)

	if len(dist) == 1 {
		return nil, errors.New("brave base accepts image names in the format NAME/VERSION/ARCH. See https://images.linuxcontainers.org for a list of supported images")
	}

	version := strings.SplitN(dist[1], ".", 2)
	distroVersion := version[0]

	if len(version) > 1 {
		distroVersion = strings.Join(version[:], "")
	}

	service := "brave-base-" + dist[0] + "-" + distroVersion

	baseConfig = BRAVEFILE

	nameRegexp, _ := regexp.Compile("<name>")
	serviceRegexp, _ := regexp.Compile("<service>")

	baseConfig = nameRegexp.ReplaceAllString(baseConfig, name)
	baseConfig = serviceRegexp.ReplaceAllString(baseConfig, service)

	err := yaml.Unmarshal([]byte(baseConfig), &bravefile)
	if err != nil {
		return nil, err
	}

	return &bravefile, nil
}

// CopyFile util function
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// WalkMatch ..
func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("Failed to get file info: " + err.Error())
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return errors.New("Failed to match filepath: " + err.Error())
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, errors.New("Failed to walk file path: " + err.Error())
	}
	return matches, nil
}

// StringSliceSearch searches a string slice for an expression and returns its indeces
func StringSliceSearch(array []string, expression string) ([]int, error) {
	pattern := regexp.MustCompile(expression)
	var result []int

	for _, s := range array {
		i := pattern.FindStringIndex(s)
		if i != nil {
			result = append(result, i[0])
		} else {
			result = append(result, -1)
		}
	}

	return result, nil
}

// StringInSlice checks if string is present in a slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// TruncateStringLeft to specific number of characters fro the left
func TruncateStringLeft(str string, num int) string {
	res := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		res = str[0:num] + "..."
	}
	return res
}

// TruncateStringRight to specific number of characters fro the right
func TruncateStringRight(str string, num int) string {
	res := str
	if len(str) > num {
		if num < 3 {
			num += 3
		}
		res = "..." + str[len(str)-(num-3):]
	}
	return res
}

// RandomSequence generates a random sequence with length n
func RandomSequence(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// FormatByteCountSI Returns formatted byte
func FormatByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// SizeCountToInt convert size strings to integers
func SizeCountToInt(s string) (int64, error) {
	unitMap := map[string]int64{
		"B":  1,
		"KB": 1000,
		"MB": 1000000,
		"GB": 1000000000,
		"TB": 1000000000000,
	}

	split := make([]string, 0)
	for i, r := range s {
		if !unicode.IsDigit(r) {
			split = append(split, strings.TrimSpace(string(s[:i])))
			split = append(split, strings.TrimSpace(string(s[i:])))
			break
		}
	}

	unit, ok := unitMap[strings.ToUpper(split[1])]
	if !ok {
		return 0, errors.New("Unrecognized size suffix " + split[1])

	}

	value, err := strconv.ParseInt(split[0], 0, 64)
	if err != nil {
		return 0, errors.New("Unable to parse " + split[0] + ": " + err.Error())
	}

	value = value * unit

	return value, nil

}

// FileHash creates MD5 for a given file
func FileHash(filePath string) (string, error) {
	operation := Info("Getting hash")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " " + operation
	s.Start()
	var MD5String string

	file, err := os.Open(filePath)
	if err != nil {
		return MD5String, errors.New("Failed to open disk file: " + err.Error())
	}

	defer file.Close()
	hash := md5.New()

	//Copy the file in the hash interface and check for errors
	if _, err := io.Copy(hash, file); err != nil {
		return MD5String, errors.New("Failed to copy a file: " + err.Error())
	}

	hashInBytes := hash.Sum(nil)[:16]
	MD5String = hex.EncodeToString(hashInBytes)

	s.Stop()

	return MD5String, nil
}

//CheckPath checks if path exists
func CheckPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//FileExists checks if path exists and ensures that it's a file
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ReadFile ..
func ReadFile(path string) (*bytes.Buffer, error) {
	filerc, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer filerc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)
	//contents := buf.String()

	return buf, nil
}

// CreateDirectory creates a directory path if not exists
func CreateDirectory(dirPath string) error {
	pathExists, err := CheckPath(dirPath)
	if err != nil {
		return err
	}
	if pathExists == false {
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
