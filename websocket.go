package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketMessage struct {
	MsgType string `json:"type"`
}

type HostListMessage struct {
	WebsocketMessage
	Hosts []Host `json:"data"`
}

func makeWebsocketHandler(upgrader *websocket.Upgrader, hosts []Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		sendHostList := true

		for {
			if sendHostList {
				sendHostList = false
				msg := HostListMessage{WebsocketMessage: WebsocketMessage{MsgType: "list"}, Hosts: hosts}
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
					return
				}
				continue
			}

			err := conn.Close()
			if err != nil {
				log.Println(err)
				return
			}
			return

		}
	}
}
