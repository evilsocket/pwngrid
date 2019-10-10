package main

import (
	"flag"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/joho/godotenv"
)

var (
	debug    = false
	routes   = false
	address  = "0.0.0.0:8666"
	env      = ".env"
	keysPath = ""
	keys     = (*crypto.KeyPair)(nil)
)

func init() {
	flag.BoolVar(&debug, "debug", debug, "Enable debug logs.")
	flag.BoolVar(&routes, "routes", routes, "Generate routes documentation.")
	flag.StringVar(&log.Output, "log", log.Output, "Log file path or empty for standard output.")
	flag.StringVar(&address, "address", address, "API address.")
	flag.StringVar(&env, "env", env, "Load .env from.")

	flag.StringVar(&keysPath, "keys", keysPath, "If set, will load RSA keys from this folder and start in peer mode.")
}

func main() {
	var err error

	flag.Parse()

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
		if keys, err = crypto.Load(keysPath); err != nil {
			log.Fatal("error while loading keys from %s: %v", keysPath, err)
		}
	}

	log.Info("pwngrid v%s starting in %s mode ...", api.Version, mode)

	if err := godotenv.Load(env); err != nil {
		log.Fatal("%v", err)
	}

	if keys == nil {
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
