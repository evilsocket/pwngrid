package utils

import (
	"github.com/evilsocket/islazy/str"
	"os/exec"
)

func Exec(executable string, args []string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}

	raw, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return "", err
	} else {
		return str.Trim(string(raw)), nil
	}
}
