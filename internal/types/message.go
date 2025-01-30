package types

import (
	"encoding/json"
)

type Message struct {
	From    string
	To      string
	Content string
}

type MessageCodec struct{}

func (c *MessageCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *MessageCodec) Decode(data []byte) (interface{}, error) {
	var m Message
	return &m, json.Unmarshal(data, &m)
}

type MessageListCodec struct{}

func (c *MessageListCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *MessageListCodec) Decode(data []byte) (interface{}, error) {
	var m []Message
	err := json.Unmarshal(data, &m)
	return m, err
}
