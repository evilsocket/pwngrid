package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/pwngrid/api"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"time"
)

var (
	debug    = false
	routes   = false
	ver      = false
	wait     = false
	inbox    = false
	del      = false
	unread   = false
	receiver = ""
	message  = ""
	page     = 1
	id       = 0
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

	flag.BoolVar(&inbox, "inbox", inbox, "Show inbox.")
	flag.StringVar(&receiver, "send", receiver, "Receiver unit fingerprint.")
	flag.StringVar(&message, "message", message, "Message body or file path if prefixed by @.")
	flag.BoolVar(&del, "delete", del, "Delete the specified message.")
	flag.BoolVar(&unread, "unread", unread, "Unread the specified message.")
	flag.IntVar(&page, "page", page, "Inbox page.")
	flag.IntVar(&id, "id", id, "Message id.")
}

func showInbox(box map[string]interface{}) {
	messages := box["messages"].([]interface{})
	numMessages := len(messages)

	if numMessages > 0 {
		records := box["records"].(float64)
		pages := box["pages"].(float64)
		columns := []string{
			"ID",
			"Date",
			"Sender",
		}
		rows := [][]string{}

		for _, m := range messages {
			var row []string
			msg := m.(map[string]interface{})

			row = []string{
				fmt.Sprintf("%d", int(msg["id"].(float64))),
				msg["created_at"].(string),
				fmt.Sprintf("%s@%s", msg["sender_name"], msg["sender"]),
			}

			if msg["seen_at"] == nil {
				for i := range row {
					row[i] = tui.Bold(row[i])
				}
			}

			rows = append(rows, row)
		}

		fmt.Println()
		tui.Table(os.Stdout, columns, rows)
		fmt.Println()

		fmt.Printf("%d of %d (page %d of %d)", numMessages, int(records), page, int(pages))
	} else {
		fmt.Println()
		fmt.Println(tui.Dim("Inbox is empty."))
	}

	fmt.Println()
}

func showMessage(msg map[string]interface{}) {
	fmt.Println(msg)
	fmt.Printf("Message from %s@%s of the %s\n\n", msg["sender_name"], msg["sender"], msg["created_at"])
	fmt.Printf("%s\n", msg["data"])
	fmt.Println()
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

	if (inbox || receiver != "") && keysPath == "" {
		keysPath = "/etc/pwnagotchi/"
	}

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

	if keys != nil {
		if inbox {
			if id == 0 {
				log.Info("fetching inbox ...")
				if box, err := server.Client.Inbox(page); err != nil {
					log.Fatal("%v", err)
				} else {
					showInbox(box)
				}
			} else if del {
				log.Info("deleting message %d ...", id)
				if _, err := server.Client.MarkInboxMessage(id, "deleted"); err != nil {
					log.Fatal("%v", err)
				}
			} else if unread {
				log.Info("marking message %d as unread ...", id)
				if _, err := server.Client.MarkInboxMessage(id, "unseen"); err != nil {
					log.Fatal("%v", err)
				}
			} else {
				log.Info("fetching message %d ...", id)

				if msg, status, err := server.InboxMessage(id); err != nil {
					log.Fatal("%d %v", status, err)
				} else {
					showMessage(msg)
					_, _ = server.Client.MarkInboxMessage(id, "seen")
				}
			}
		} else if receiver != "" {
			if message == "" {
				log.Fatal("-message can not be empty")
			} else if message[0] == '@' {
				if message, err := ioutil.ReadFile(message[1:]); err != nil {
					log.Fatal("error reading %s: %v", message[1:], err)
				}
			}

			if status, err := server.SendMessage(receiver, []byte(message)); err != nil {
				log.Fatal("%d %v", status, err)
			} else {
				log.Info("message sent")
			}
		}
	} else {
		server.Run(address)
	}
}
