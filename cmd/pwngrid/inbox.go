package main

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/tui"
	"github.com/evilsocket/pwngrid/api"
	"io/ioutil"
	"os"
)

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
		}
	} else {
		fmt.Println()
		fmt.Println(tui.Dim("Inbox is empty."))
	}

	fmt.Println()
}

func showMessage(msg map[string]interface{}) {
	fmt.Println()
	fmt.Printf("Message from %s@%s of the %s\n\n", msg["sender_name"], msg["sender"], msg["created_at"])
	if output == "" {
		fmt.Printf("%s\n", msg["data"])
		fmt.Println()
	} else if err := ioutil.WriteFile(output, msg["data"].([]byte), os.ModePerm); err != nil {
		log.Fatal("error writing to %s: %v", output, err)
	} else {
		log.Info("%s written", output)
	}
}

func doInbox(server *api.API) {
	var err error

	if inbox {
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
		os.Exit(0)
	} else if receiver != "" {
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
		os.Exit(0)
	}
}