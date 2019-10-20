package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/version"
)

func main() {
	flag.Parse()

	setupCore()
	defer cleanup()

	// just print the version and exit
	if ver {
		fmt.Println(version.Version)
		return
	}

	// from here on we need logging
	if err := log.Open(); err != nil {
		panic(err)
	}
	defer log.Close()

	// do mode related initialization
	setupMode()

	// if we're in peer mode and is an inbox action
	if inbox {
		inboxMain()
	} else {
		// just start the API in either modes
		server.Run(address)
	}
}
