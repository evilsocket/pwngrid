package utils

import (
	"github.com/evilsocket/islazy/log"
	"os"
	"strings"
)

func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Warning("%v", err)
		return ""
	}

	if strings.HasSuffix(name, ".local") {
		name = strings.Replace(name, ".local", "", -1)
	}

	return name
}
