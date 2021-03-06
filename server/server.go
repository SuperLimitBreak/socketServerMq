package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SuperLimitBreak/channelMq"
	"github.com/SuperLimitBreak/socketServerMq/server/connectionManager"
	tSockConn "github.com/SuperLimitBreak/socketServerMq/server/connectionTypes/tcpSocket"
	wSockConn "github.com/SuperLimitBreak/socketServerMq/server/connectionTypes/websocket"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
)

var upgrader websocket.Upgrader
var mq *channelMq.MQ
var connMan *connectionManager.ConnectionManager

func init() {
	log.SetLevel(log.DebugLevel)

	mq = channelMq.NewMQ()
	mq.EmptyKeyGets(channelMq.ALL_MESSAGES)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connMan = connectionManager.NewConnectionManager(mq)
}

func StartAll() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go startWs(wg)
	go startTs(wg)

	wg.Wait()
}

func startTs(wg *sync.WaitGroup) {
	log.Info("Starting tcpSocket server")

	ln, err := net.Listen("tcp", ":9872")
	if err != nil {
		log.WithError(err).Fatal("Failed to start TCP server")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.WithError(err).Fatal("TCP accept loop failed")
		}

		go serveTs(conn)
	}

	wg.Done()
}

func serveTs(conn net.Conn) {
	log.Debug("New tcp connection")

	tsc := tSockConn.NewTcpSocketConn(conn)
	connMan.AddConnection(tsc)
	tsc.Start()
}

func startWs(wg *sync.WaitGroup) {
	http.HandleFunc("/ws", serveWs)

	log.Info("Starting Websocket server")

	err := http.ListenAndServe(":9873", nil)
	if err != nil {
		log.WithError(err).Fatal("Failed to serve websockets")
	}

	wg.Done()
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Debug("New Websocket Request")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Failed to upgrade HTTP handler")
		return
	}

	log.Debug("New websocket connection (upgrade ok)")

	wsc := wSockConn.NewWebsocketConn(ws)
	connMan.AddConnection(wsc)
	wsc.Start()
}
