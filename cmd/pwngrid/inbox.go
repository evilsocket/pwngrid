package main

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/pwngrid/api"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func clearScreen() {
	var what []string
	if runtime.GOOS == "windows" {
		what = []string{"cmd", "/c", "cls"}
	} else {
		what = []string{"clear", ""}
	}
	cmd := exec.Command(what[0], what[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func showInbox(server *api.API, box map[string]interface{}) {
	messages := box["messages"].([]interface{})
	numMessages := len(messages)

	if numMessages > 0 {
		if clear {
			log.Info("clearing %d messages", numMessages)
			for _, m := range messages {
				msg := m.(map[string]interface{})
				msgID := int(msg["id"].(float64))
				log.Info("deleting message %d ...", msgID)
				if _, err := server.Client.MarkInboxMessage(msgID, "deleted"); err != nil {
					log.Error("%v", err)
				}
			}
		} else {
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

				t, err := time.Parse(time.RFC3339, msg["created_at"].(string))
				if err != nil {
					panic(err)
				}

				row = []string{
					fmt.Sprintf("%d", int(msg["id"].(float64))),
					t.Format("02 January 2006, 3:04 PM"),
					fmt.Sprintf("%s@%s", msg["sender_name"], msg["sender"]),
				}

				if msg["seen_at"] != nil {
					for i := range row {
						row[i] = tui.Dim(row[i])
					}
				}

				rows = append(rows, row)
			}

			fmt.Println()
			tui.Table(os.Stdout, columns, rows)
			fmt.Println()

			fmt.Printf("%d of %d (page %d of %d)", numMessages, int(records), page, int(pages))
		}
	} else {
		fmt.Println()
		fmt.Println(tui.Dim("Inbox is empty."))
	}

	fmt.Println()
}

func showMessage(msg map[string]interface{}) {
	t, err := time.Parse(time.RFC3339, msg["created_at"].(string))
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Printf("From: %s@%s\n", msg["sender_name"], msg["sender"])
	fmt.Printf("Date: %s\n\n", t.Format("02 January 2006, 3:04 PM"))
	if output == "" {
		fmt.Printf("%s\n", msg["data"])
		fmt.Println()
	} else if err := ioutil.WriteFile(output, msg["data"].([]byte), os.ModePerm); err != nil {
		log.Fatal("error writing to %s: %v", output, err)
	} else {
		log.Info("%s written", output)
	}
}

func sendMessage() {
	var err error

	// send a message
	var raw []byte
	if message == "" {
		log.Fatal("-message can not be empty")
	} else if message[0] == '@' {
		log.Info("reading %s ...", message[1:])
		if raw, err = ioutil.ReadFile(message[1:]); err != nil {
			log.Fatal("error reading %s: %v", message[1:], err)
		}
	} else {
		raw = []byte(message)
	}

	if status, err := server.SendMessage(receiver, raw); err != nil {
		log.Fatal("%d %v", status, err)
	} else {
		log.Info("message sent")
	}
}

func doInbox(server *api.API) {
	if receiver != "" {
		sendMessage()
	} else if inbox {
		// just show the inbox
		if id == 0 {
			log.Info("fetching inbox ...")
			if box, err := server.Client.Inbox(page); err != nil {
				log.Fatal("%v", err)
			} else {
				showInbox(server, box)
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
	}
}

func inboxMain() {
	if inbox {
		doInbox(server)
		if loop {
			ticker := time.NewTicker(time.Duration(loopPeriod) * time.Second)
			for _ = range ticker.C {
				clearScreen()
				doInbox(server)
			}
		}
		os.Exit(0)
	}
}