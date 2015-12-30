package schema

import (
	"encoding/json"
)

const (
	MESSAGE_ACTION   string = `message`
	SUBSCRIBE_ACTION string = `subscribe`
)

// Valid data sent to the server is a top level json object with 2 keys.
// These keys are "action" and "data"
//
// Valid data for "action" is either the string message or the string subscribe
//
// Valid data for data is either an array of strings for action subscribe, to
// denote the group(s) to subscribe to, or an array of json objects if the
// action is message.
//
// These objects my optionaly contain the key deviceId who's value is a string,
// this is used to denote the group to send the message to.

type GenericMessage struct {
	Action string           `json:action`
	Data   *json.RawMessage `json:data`
}

func (g GenericMessage) Keys() ([]string, error) {
	var keys []string
	err := json.Unmarshal(*g.Data, &keys)

	return keys, err
}

func (g GenericMessage) Messages() ([]Message, error) {
	rtn := []Message{}
	var msgs []*json.RawMessage

	err := json.Unmarshal(*g.Data, &msgs)
	if err != nil {
		return rtn, err
	}

	for i, _ := range msgs {
		m := Message{msgs[i]}
		rtn = append(rtn, m)
	}
	return rtn, err
}

func (g GenericMessage) IsAction(term string) bool {
	return g.Action == term
}

type Message struct {
	Data *json.RawMessage
}

func (m Message) Key() (*string, error) {
	var partial map[string]*json.RawMessage

	err := json.Unmarshal(*m.Data, &partial)
	if err != nil {
		return nil, err
	}

	devId, ok := partial["deviceid"]
	if !ok {
		return nil, nil
	}

	var key string
	err = json.Unmarshal(*devId, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func WrapWithTopLevelMessage(msg []byte) []byte {
	front := []byte(`{"action":"message","data":[`)
	end := []byte(`]}`)

	rtn := append(front, msg...)
	rtn = append(rtn, end...)

	return rtn
}
