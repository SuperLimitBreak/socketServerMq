package connectionManager

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/SuperLimitBreak/channelMq"
	"github.com/SuperLimitBreak/socketServerMq/server/schema"
)

type connectionInterface interface {
	GetIngressChan() chan []byte
	GetEgressChan() chan []byte
}

func newConnection(conn connectionInterface) *connection {
	return &connection{
		conn.GetIngressChan(),
		conn.GetEgressChan(),
	}
}

type connection struct {
	ingressChan chan []byte
	egressChan  chan []byte
}

func (c *connection) processChan(ch chan []byte) {
	defer func() {
		recover()
	}()

	log.Debug("started processChan worker")
	for {
		msg, ok := <-ch
		if !ok {
			log.Debug("conection channel closed")
			return
		}

		log.Debug("Sending data to egressChan")
		c.egressChan <- msg
	}
}

func (c *connection) work(mq *channelMq.MQ) {
	log.Debug("Doing First subscribe")
	mqChan := mq.Subscribe([]string{})
	go c.processChan(mqChan)

	for {
		log.Debug("Waiting on ingressChan")
		msg, ok := <-c.ingressChan
		if !ok {
			log.Debug("conn closed bailing")
			close(c.egressChan)
			return
		}

		log.Debug("New message")

		//Process message
		var generic schema.GenericMessage

		err := json.Unmarshal(msg, &generic)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"json": string(msg),
			}).Warn("malformed Json in ingress queue")
			continue
		}

		//switch on message type
		switch {

		// message action
		case generic.IsAction(schema.MESSAGE_ACTION):
			log.Debug("New message action")

			messages, err := generic.Messages()
			if err != nil {
				log.WithError(err).Warn("Failed to get messages")
				continue
			}

			for _, msg := range messages {
				key, err := msg.Key()
				if err != nil {
					log.WithError(err).Warn("Failed to get key")
					continue
				}

				d := schema.WrapWithTopLevelMessage(
					*msg.Data,
				)

				log.Debugf("sending %v", string(d))
				if key == nil {
					mq.Send(d, []string{})
				} else {
					mq.Send(d, []string{*key})
				}
			}

		// subscribe action
		case generic.IsAction(schema.SUBSCRIBE_ACTION):
			log.Debug("New debug action")

			keys, err := generic.Keys()
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"json": string(msg),
				}).Warn("malformed Json subscribe in ingress queue")
				continue
			}

			log.Debug("Unsubscribe from mq")
			mq.Unsubscribe(mqChan)

			log.Debug("close mq channel")
			close(mqChan)

			log.Debug("replace mq channel with new channel")
			mqChan = mq.Subscribe(keys)

			log.Debug("Start new worker")
			go c.processChan(mqChan)

		default:
			log.WithFields(log.Fields{
				"json":   string(msg),
				"action": generic.Action,
			}).Warn("Unknown action in json")
			continue
		}

	}
}
