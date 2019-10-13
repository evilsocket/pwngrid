package main

import (
	"flag"
	"fmt"
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

var (
	debug      = false
	routes     = false
	ver        = false
	wait       = false
	inbox      = false
	del        = false
	unread     = false
	clear      = false
	receiver   = ""
	message    = ""
	output     = ""
	page       = 1
	id         = 0
	address    = "0.0.0.0:8666"
	env        = ".env"
	iface      = "mon0"
	keysPath   = ""
	keys       = (*crypto.KeyPair)(nil)
	peer       = (*mesh.Peer)(nil)
	cpuProfile = ""
	memProfile = ""
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

	flag.StringVar(&iface, "iface", iface, "Monitor interface to use for mesh advertising.")
	flag.IntVar(&mesh.SignalingPeriod, "signaling-period", mesh.SignalingPeriod, "Period in milliseconds for mesh signaling frames.")

	flag.BoolVar(&inbox, "inbox", inbox, "Show inbox.")
	flag.StringVar(&receiver, "send", receiver, "Receiver unit fingerprint.")
	flag.StringVar(&message, "message", message, "Message body or file path if prefixed by @.")
	flag.StringVar(&output, "output", output, "Write message body to this file instead of the standard output.")
	flag.BoolVar(&del, "delete", del, "Delete the specified message.")
	flag.BoolVar(&unread, "unread", unread, "Unread the specified message.")
	flag.BoolVar(&clear, "clear", unread, "Delete all messages of the given page of the inbox.")
	flag.IntVar(&page, "page", page, "Inbox page.")
	flag.IntVar(&id, "id", id, "Message id.")

	flag.StringVar(&cpuProfile, "cpu-profile", cpuProfile, "Generate CPU profile to this file.")
	flag.StringVar(&memProfile, "mem-profile", cpuProfile, "Generate memory profile to this file.")
}

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
		pprof.WriteHeapProfile(f)
		f.Close()
	}
}

func main() {
	var err error

	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
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
		pprof.StartCPUProfile(f)
	}

	defer cleanup()

	if ver {
		fmt.Println(version.Version)
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

	if (inbox || receiver != "") && keysPath == "" {
		keysPath = "/etc/pwnagotchi/"
	}

	mode := "server"
	if keysPath != "" {
		mode = "peer"
	}

	log.Info("pwngrid v%s starting in %s mode ...", version.Version, mode)

	if mode == "peer" {
		// wait for keys to be generated
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

		peer = mesh.MakeLocalPeer(utils.Hostname(), keys)
		if err = peer.StartAdvertising(iface); err != nil {
			log.Fatal("error while starting signaling: %v", err)
		}

		if router, err := mesh.StartRouting(iface, peer); err != nil {
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

	if keys == nil {
		if err := godotenv.Load(env); err != nil {
			log.Fatal("%v", err)
		}

		if err := models.Setup(); err != nil {
			log.Fatal("%v", err)
		}
	}

	err, server := api.Setup(keys, peer, routes)
	if err != nil {
		log.Fatal("%v", err)
	}

	if keys != nil {
		doInbox(server)
	}

	server.Run(address)
}
