package main

import (
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/mesh"
	"github.com/evilsocket/pwngrid/models"
	"github.com/evilsocket/pwngrid/utils"
	"github.com/evilsocket/pwngrid/version"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"
)

func cleanup() {
	if cpuProfile != "" {
		log.Info("writing CPU profile to %s ...", cpuProfile)
		pprof.StopCPUProfile()
	}

	if memProfile != "" {
		log.Info("writing memory profile to %s ...", memProfile)
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal("%v", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
	}
}

func setupCore() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Warning("received signal %v", sig)
			cleanup()
			os.Exit(0)
		}
	}()

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
	}

	if debug {
		log.Level = log.DEBUG
	} else {
		log.Level = log.INFO
	}
	log.OnFatal = log.ExitOnFatal
}

func waitForKeys() {
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

func setupMesh() {
	var err error
	peer = mesh.MakeLocalPeer(utils.Hostname(), keys)
	if err = peer.StartAdvertising(iface); err != nil {
		log.Fatal("error while starting signaling: %v", err)
	}
	if router, err := mesh.StartRouting(iface, peersPath, peer); err != nil {
		log.Fatal("%v", err)
	} else {
		router.OnNewPeer(func(ident string, peer *mesh.Peer) {
			log.Info("detected new peer %s on channel %d", peer.ID(), peer.Channel)
		})
		router.OnPeerLost(func(ident string, peer *mesh.Peer) {
			log.Info("peer %s lost (inactive for %fs)", peer.ID(), peer.InactiveFor())
		})
	}
	log.Info("peer %s signaling is ready", peer.ID())
}

func setupDB() {
	if err := godotenv.Load(env); err != nil {
		log.Fatal("%v", err)
	}
	if err := models.Setup(); err != nil {
		log.Fatal("%v", err)
	}
}

func setupMode() string {
	var err error

	// in case -inbox was not explicitly passed
	if receiver != "" || loop == true || id > 0 {
		inbox = true
	}

	// for inbox actions, set the keys to the default path if empty
	if (whoami || inbox) && keysPath == "" {
		keysPath = "/etc/pwnagotchi/"
	}

	// generate keypair
	if generate {
		if keysPath == "" {
			log.Fatal("no -keys path specified")
		} else if crypto.KeysExist(keysPath) {
			log.Fatal("keypair already exists in %s", keysPath)
		}

		if _, err = crypto.LoadOrCreate(keysPath, 4096); err != nil {
			log.Fatal("error generating RSA keypair: %v", err)
		} else {
			log.Info("keypair saved to %s", keysPath)
		}
		os.Exit(0)
	}

	mode := "server"
	// if keys have been passed explicitly, or one of the inbox actions
	// has been specified, we're running on the unit
	if keysPath != "" {
		mode = "peer"
	}

	log.Info("pwngrid v%s starting in %s mode ...", version.Version, mode)

	if mode == "peer" {
		// wait for keys to be generated
		if wait {
			waitForKeys()
		}
		// load the keys
		if keys, err = crypto.Load(keysPath); err != nil {
			log.Fatal("error while loading keys from %s: %v", keysPath, err)
		}
		// print identity and exit
		if whoami {
			log.Info("https://pwnagotchi.ai/pwnfile/#!%s", keys.FingerprintHex)
			os.Exit(0)
		}
		// only start mesh signaling if this is not an inbox action
		if !inbox {
			setupMesh()
		}
	} else if mode == "server" {
		// server side we just need to setup the database connection
		setupDB()
	}

	// setup the proper routes for either server or peer mode
	err, server = api.Setup(keys, peer)
	if err != nil {
		log.Fatal("%v", err)
	}

	return mode
}
