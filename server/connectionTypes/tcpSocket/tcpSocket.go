package tcpSocket

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
)

func NewTcpSocketConn(c net.Conn) *TcpSocketConn {
	tsc := &TcpSocketConn{
		c,
		make(chan []byte, 256),
		make(chan []byte, 256),
		make(chan struct{}),
	}

	return tsc
}

type TcpSocketConn struct {
	conn        net.Conn
	ingressChan chan []byte
	egressChan  chan []byte
	stop        chan struct{}
}

func (c *TcpSocketConn) GetIngressChan() chan []byte {
	return c.ingressChan
}

func (c *TcpSocketConn) GetEgressChan() chan []byte {
	return c.egressChan
}

func (c *TcpSocketConn) Start() {
	log.Debug("starting ingress/egress worker")
	go c.ingress()
	go c.egress()
}

func (c *TcpSocketConn) ingress() {
	for {
		select {
		case <-c.stop:
			log.Debug("ingress loop stop")
			close(c.ingressChan)
			return

		default:
			message, err := bufio.NewReader(c.conn).ReadBytes('\n')
			if err != nil {
				log.WithError(err).Error("failed to read from tcp socket")
				c.Close()

				close(c.ingressChan)
				return
			}

			log.Debug("Send message to ingressChan")
			c.ingressChan <- message

		}
	}
}

func (c *TcpSocketConn) egress() {
	for {
		select {
		case <-c.stop:
			return

		case msg := <-c.egressChan:
			str := string(msg)

			n, err := fmt.Fprintf(c.conn, str+"\n")

			if (err != nil) || (n < len(str)) {
				log.WithError(err).WithFields(log.Fields{
					"length":  len(str),
					"written": n,
				}).Error("failed to send to tcp socket")

				c.Close()
				return
			}
		}
	}
}

func (c *TcpSocketConn) closeStop() {
	defer func() {
		if r := recover(); r != nil {
			// here to prevent close on closed chan race panic
		}
	}()

	close(c.stop)
}

func (c *TcpSocketConn) Close() error {
	c.closeStop()
	return c.conn.Close()
}
