package websocket

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeWait      = time.Second
	pongWait       = 500 * time.Millisecond
	pingPeriod     = (pongWait * 9) / 10 // Must be < pongWait
	maxMessageSize = 4096
)

func NewWebsocketConn(ws *websocket.Conn) *WebsocketConn {
	wsc := &WebsocketConn{
		ws,
		make(chan []byte, 256),
		make(chan []byte, 256),
		make(chan struct{}),
	}

	//init
	wsc.ws.SetReadLimit(maxMessageSize)
	wsc.ws.SetPongHandler(wsc.pongHandle)

	return wsc
}

type WebsocketConn struct {
	ws          *websocket.Conn
	ingressChan chan []byte
	egressChan  chan []byte
	stop        chan struct{}
}

func (c *WebsocketConn) Start() {
	go c.pingLoop()
	go c.ingress()
	go c.egress()
}

func (c *WebsocketConn) GetIngressChan() chan []byte {
	return c.ingressChan
}

func (c *WebsocketConn) GetEgressChan() chan []byte {
	return c.egressChan
}

func (c *WebsocketConn) pingLoop() {
	ticker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-c.stop:
			return

		case <-ticker.C:
			c.write(websocket.PingMessage, []byte{})
		}
	}
}

func (c *WebsocketConn) pongHandle(string) error {
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	return nil
}

func (c *WebsocketConn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *WebsocketConn) ingress() {
	for {
		select {
		case <-c.stop:
			close(c.ingressChan)
			return

		default:
			c.ws.SetReadDeadline(time.Now().Add(pongWait))
			_, message, err := c.ws.ReadMessage()
			if err != nil {
				log.WithError(err).Error("failed to read from websocket")
				c.Close()
				close(c.ingressChan)
				return
			}

			c.ingressChan <- message
		}
	}
}

func (c *WebsocketConn) egress() {
	for {
		select {
		case <-c.stop:
			return

		case msg := <-c.egressChan:
			if err := c.write(websocket.TextMessage, msg); err != nil {
				log.WithError(err).Error("failed to send to websocket")
				c.Close()
				return
			}
		}
	}
}

func (c *WebsocketConn) closeStop() {
	defer func() {
		if r := recover(); r != nil {
			// here to prevent close on closed chan race panic
		}
	}()

	close(c.stop)
}

func (c *WebsocketConn) Close() error {
	c.closeStop()

	c.write(websocket.CloseMessage, []byte{})
	return c.ws.Close()
}
