package shared

import (
	"os"
	"os/exec"
)

// ExecCommand runs a command
func ExecCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	return err
}

// ExecCommandWReturn ..
func ExecCommandWReturn(name string, arg ...string) (string, error) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
