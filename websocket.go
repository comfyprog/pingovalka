package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketMessage struct {
	MsgType string `json:"type"`
}

type HostListMessage struct {
	WebsocketMessage
	Hosts []Host `json:"data"`
}

type HostStatusMessage struct {
	WebsocketMessage
	Host Host `json:"data"`
}

func makeWebsocketHandler(upgrader *websocket.Upgrader, pingMux *PingMux, socketPingInterval time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		defer func() {
			err = conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
			if err != nil {
				log.Println(err)
			}
			err = conn.Close()
			if err != nil {
				log.Println(err)
			}
		}()

		pingChan, pingChanCleanFunc := pingMux.AddSubscriber()
		defer pingChanCleanFunc()

		pingTicker := time.NewTicker(socketPingInterval)
		defer pingTicker.Stop()

		hosts := pingMux.GetHosts()
		msg := HostListMessage{WebsocketMessage: WebsocketMessage{MsgType: "list"}, Hosts: hosts}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println(err)
			return
		}

		for {
			select {
			case host, ok := <-pingChan:
				if !ok {
					return
				}
				msg := HostStatusMessage{WebsocketMessage: WebsocketMessage{MsgType: "status"}, Host: host}
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
					return
				}
				pingTicker.Reset(socketPingInterval)

			case <-pingTicker.C:
				err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
