package main

import (
	"crypto/tls"
	"log"
	"net"
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

func makeWebsocketHandler(upgrader *websocket.Upgrader, pingMux *PingMux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		err = conn.UnderlyingConn().(*tls.Conn).NetConn().(*net.TCPConn).SetKeepAlive(true)
		if err != nil {
			log.Println(err)
			return
		}

		pingChan, pingChanCleanFunc := pingMux.AddSubscriber()
		defer pingChanCleanFunc()

		sendHostList := true

		for {
			if sendHostList {
				sendHostList = false
				hosts := pingMux.GetHosts()
				msg := HostListMessage{WebsocketMessage: WebsocketMessage{MsgType: "list"}, Hosts: hosts}
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
					return
				}
				continue
			}

			for host := range pingChan {
				msg := HostStatusMessage{WebsocketMessage: WebsocketMessage{MsgType: "status"}, Host: host}
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
					return
				}
			}

			conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
			err := conn.Close()
			if err != nil {
				log.Println(err)
				return
			}
			return

		}
	}
}
