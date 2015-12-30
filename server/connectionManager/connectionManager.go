package connectionManager

import (
	log "github.com/Sirupsen/logrus"
	"github.com/SuperLimitBreak/channelMq"
)

func NewConnectionManager(mq *channelMq.MQ) *ConnectionManager {
	return &ConnectionManager{
		mq,
	}
}

type ConnectionManager struct {
	mq *channelMq.MQ
}

func (c *ConnectionManager) AddConnection(ci connectionInterface) {
	log.Debug("Adding new connection")
	conn := newConnection(ci)

	log.Debug("Starting connection worker")
	go conn.work(c.mq)
}
