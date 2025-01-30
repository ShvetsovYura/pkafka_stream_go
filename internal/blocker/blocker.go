package blocker

import (
	"context"
	"slices"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/types"
	"github.com/lovoo/goka"
)

type Blocker struct {
	brokers       []string
	group         goka.Group
	table         goka.Table
	blockerStream goka.Stream
}

func NewBlocker(brokers []string, blockerGroupName string, blockerInputTopicName string) *Blocker {
	g := goka.Group(blockerGroupName)
	return &Blocker{
		brokers:       brokers,
		group:         g,
		table:         goka.GroupTable(g),
		blockerStream: goka.Stream(blockerInputTopicName),
	}
}

func block(ctx goka.Context, msg interface{}) {
	var s *types.BlockValue
	if v := ctx.Value(); v == nil {
		s = new(types.BlockValue)
	} else {
		s = v.(*types.BlockValue)
	}
	val, ok := msg.(*types.BlockEvent)
	if ok {
		if val.IsBlocked {
			if !slices.Contains(s.BlockedByUsers, val.User) {
				s.BlockedByUsers = append(s.BlockedByUsers, val.User)
			}
		} else {
			if slices.Contains(s.BlockedByUsers, val.User) {
				idx := slices.Index(s.BlockedByUsers, val.User)
				if idx > -1 {
					s.BlockedByUsers = append(s.BlockedByUsers[:idx], s.BlockedByUsers[idx+1:]...)
				}
			}
		}
	}

	ctx.SetValue(s)
}

func (b *Blocker) Run(ctx context.Context) func() error {
	return func() error {
		g := goka.DefineGroup(b.group,
			goka.Input(b.blockerStream, new(types.BlockEventCodec), block),
			goka.Persist(new(types.BlockValueCodec)),
		)
		p, err := goka.NewProcessor(b.brokers, g)
		if err != nil {
			return err
		}

		return p.Run(ctx)
	}
}
