package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SuperLimitBreak/channelMq"
	"github.com/SuperLimitBreak/socketServerMq/connectionManager"
	wSockConn "github.com/SuperLimitBreak/socketServerMq/connectionTypes/websocket"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader websocket.Upgrader
var mq *channelMq.MQ
var connMan *connectionManager.ConnectionManager

func init() {
	log.Info("Initializing channelMq")
	mq = channelMq.NewMQ()
	mq.EmptyKeyGets(channelMq.ALL_MESSAGES)

	log.Info("Initializing websocket upgrader")
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	log.Info("Initializing connectionManager")
	connMan = connectionManager.NewConnectionManager(mq)
}

func main() {
	startWs()
}

func startWs() {
	http.HandleFunc("/ws", serveWs)

	log.Info("Starting Websocket server...")
	http.ListenAndServe("localhost:6543", nil)
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Info("New Websocket Request")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Failed to upgrade HTTP handler")
		return
	}

	wsc := wSockConn.NewWebsocketConn(ws)
	connMan.AddConnection(wsc)
	wsc.Start()
}
