package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/joho/godotenv"
	"time"
)

var (
	debug    = false
	routes   = false
	ver      = false
	wait     = false
	address  = "0.0.0.0:8666"
	env      = ".env"
	keysPath = ""
	keys     = (*crypto.KeyPair)(nil)
)

func init() {
	flag.BoolVar(&ver, "version", ver, "Print version and exit.")
	flag.BoolVar(&debug, "debug", debug, "Enable debug logs.")
	flag.BoolVar(&routes, "routes", routes, "Generate routes documentation.")
	flag.StringVar(&log.Output, "log", log.Output, "Log file path or empty for standard output.")
	flag.StringVar(&address, "address", address, "API address.")
	flag.StringVar(&env, "env", env, "Load .env from.")

	flag.StringVar(&keysPath, "keys", keysPath, "If set, will load RSA keys from this folder and start in peer mode.")
	flag.BoolVar(&wait, "wait", wait, "Wait for keys to be generated.")
	flag.IntVar(&api.ClientTimeout, "client-timeout", api.ClientTimeout, "Timeout in seconds for requests to the server when in peer mode.")
	flag.StringVar(&api.ClientTokenFile, "client-token", api.ClientTokenFile, "File where to store the API token.")
}

func main() {
	var err error

	flag.Parse()

	if ver {
		fmt.Println(api.Version)
		return
	}

	if debug {
		log.Level = log.DEBUG
	} else {
		log.Level = log.INFO
	}
	log.OnFatal = log.ExitOnFatal

	if err := log.Open(); err != nil {
		panic(err)
	}
	defer log.Close()

	mode := "server"
	if keysPath != "" {
		mode = "peer"

		if wait {
			privPath := crypto.PrivatePath(keysPath)
			for {
				if !fs.Exists(privPath) {
					log.Debug("waiting for %s ...", privPath)
					time.Sleep(1 * time.Second)
				} else {
					// give it a moment to finish disk sync
					time.Sleep(2 * time.Second)
					log.Info("%s found", privPath)
					break
				}
			}
		}

		if keys, err = crypto.Load(keysPath); err != nil {
			log.Fatal("error while loading keys from %s: %v", keysPath, err)
		}
	}

	log.Info("pwngrid v%s starting in %s mode ...", api.Version, mode)

	if keys == nil {
		if err := godotenv.Load(env); err != nil {
			log.Fatal("%v", err)
		}

		if err := models.Setup(); err != nil {
			log.Fatal("%v", err)
		}
	}

	err, server := api.Setup(keys, routes)
	if err != nil {
		log.Fatal("%v", err)
	}

	server.Run(address)
}
