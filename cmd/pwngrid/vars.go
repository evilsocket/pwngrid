package main

import (
	"flag"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/mesh"
)

var (
	debug      = false
	ver        = false
	wait       = false
	inbox      = false
	del        = false
	unread     = false
	clear      = false
	whoami     = false
	generate   = false
	loop       = false
	nodb       = false
	loopPeriod = 30
	receiver   = ""
	message    = ""
	output     = ""
	page       = 1
	id         = 0
	address    = "0.0.0.0:8666"
	env        = ".env"
	iface      = "mon0"
	keysPath   = ""
	peersPath  = "/root/peers"
	keys       = (*crypto.KeyPair)(nil)
	router     = (*mesh.Router)(nil)
	peer       = (*mesh.Peer)(nil)
	server     = (*api.API)(nil)
	cpuProfile = ""
	memProfile = ""
)

func init() {
	flag.BoolVar(&ver, "version", ver, "Print version and exit.")
	flag.BoolVar(&debug, "debug", debug, "Enable debug logs.")
	flag.BoolVar(&nodb, "no-db", debug, "Don't fail if database connection can't be enstablished.")
	flag.StringVar(&log.Output, "log", log.Output, "Log file path or empty for standard output.")
	flag.StringVar(&address, "address", address, "API address.")
	flag.StringVar(&env, "env", env, "Load .env from.")

	flag.StringVar(&keysPath, "keys", keysPath, "If set, will load RSA keys from this folder and start in peer mode.")
	flag.BoolVar(&generate, "generate", generate, "Generate an RSA keypair if it doesn't exist yet.")
	flag.BoolVar(&wait, "wait", wait, "Wait for keys to be generated.")
	flag.IntVar(&api.ClientTimeout, "client-timeout", api.ClientTimeout, "Timeout in seconds for requests to the server when in peer mode.")
	flag.StringVar(&api.ClientTokenFile, "client-token", api.ClientTokenFile, "File where to store the API token.")

	flag.StringVar(&iface, "iface", iface, "Monitor interface to use for mesh advertising.")
	flag.StringVar(&peersPath, "peers", peersPath, "path to save historical information of met peers.")
	flag.IntVar(&mesh.SignalingPeriod, "signaling-period", mesh.SignalingPeriod, "Period in milliseconds for mesh signaling frames.")

	flag.BoolVar(&whoami, "whoami", whoami, "Prints the public key fingerprint and exit.")
	flag.BoolVar(&inbox, "inbox", inbox, "Show inbox.")
	flag.BoolVar(&loop, "loop", loop, "Keep refreshing and showing inbox.")
	flag.IntVar(&loopPeriod, "loop-period", loopPeriod, "Period in seconds to refresh the inbox.")
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
