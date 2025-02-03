package collector

import (
	"context"
	"encoding/json"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/types"
	"github.com/lovoo/goka"
)

const maxMessages = 5

type MessageListCodec struct{}

func (c *MessageListCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *MessageListCodec) Decode(data []byte) (interface{}, error) {
	var m []types.Message
	err := json.Unmarshal(data, &m)
	return m, err
}

type Collector struct {
	brokers        []string
	group          goka.Group
	receivedStream goka.Stream
}

func NewCollector(brokers []string, groupName string, receivedStream string) *Collector {
	return &Collector{
		brokers:        brokers,
		group:          goka.Group(groupName),
		receivedStream: goka.Stream(receivedStream),
	}
}

func collect(ctx goka.Context, msg interface{}) {
	var ml []types.Message
	if v := ctx.Value(); v != nil {
		ml = v.([]types.Message)
	}

	m := msg.(*types.Message)
	ml = append(ml, *m)

	if len(ml) > maxMessages {
		ml = ml[len(ml)-maxMessages:]
	}
	ctx.SetValue(ml)
}

func (c *Collector) Run(ctx context.Context) func() error {
	return func() error {
		g := goka.DefineGroup(c.group,
			goka.Input(c.receivedStream, new(types.MessageCodec), collect),
			goka.Persist(new(MessageListCodec)),
		)
		p, err := goka.NewProcessor(c.brokers, g)
		if err != nil {
			return err
		}
		return p.Run(ctx)
	}
}
