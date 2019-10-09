package main

import (
	"flag"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/models"
	"github.com/joho/godotenv"
	"os"
)

var (
	debug   = false
	address = "0.0.0.0:8666"
	env = ".env"
)

func init() {
	flag.BoolVar(&debug, "debug", debug, "Enable debug logs.")
	flag.StringVar(&log.Output, "log", log.Output, "Log file path or empty for standard output.")
	flag.StringVar(&address, "address", address, "API address.")
	flag.StringVar(&env, "env", env, "Load .env from.")
}

func main() {
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

	log.Info("pwngrid v%s starting ...", api.Version)

	if err := godotenv.Load(env); err != nil {
		log.Fatal("%v", err)
	}

	if err := models.Setup(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"),
		os.Getenv("DB_HOST"), os.Getenv("DB_NAME")); err != nil {
		log.Fatal("%v", err)
	}

	if err, server := api.Setup(); err != nil {
		log.Fatal("%v", err)
	} else {
		server.Run(address)
	}
}
