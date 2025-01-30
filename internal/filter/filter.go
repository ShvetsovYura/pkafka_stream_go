package filter

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/types"
	"github.com/lovoo/goka"
)

type Filter struct {
	brokers        []string
	in             goka.Stream
	out            goka.Stream
	filterGroup    goka.Group
	blockerTable   goka.Table
	stopwordsTable goka.Table
}

func NewFilter(brokers []string, inTopic string, outTopic string, filterGroupName string, blockerGroupName string, stopwordsGroupName string) *Filter {
	return &Filter{
		brokers:        brokers,
		in:             goka.Stream(inTopic),
		out:            goka.Stream(outTopic),
		filterGroup:    goka.Group(filterGroupName),
		blockerTable:   goka.GroupTable(goka.Group(blockerGroupName)),
		stopwordsTable: goka.GroupTable(goka.Group(stopwordsGroupName)),
	}
}

func (f *Filter) filter(ctx goka.Context, msg any) {
	fmt.Println("filter")
	m, ok := msg.(*types.Message)
	if !ok {
		fmt.Println("not type assert msg")
		return
	}
	v := ctx.Join(f.blockerTable)
	if v != nil {
		blockedByUsers := v.(*types.BlockValue).BlockedByUsers

		if slices.Contains(blockedByUsers, m.To) {
			fmt.Println("user is blocked")
			return
		}
	}

	processed_msg := f.filter_stopwords(ctx, m)
	ctx.Emit(f.out, processed_msg.To, processed_msg)
}

func (f *Filter) filter_stopwords(ctx goka.Context, m *types.Message) *types.Message {
	words := strings.Split(m.Content, " ")
	for i, w := range words {
		if tw := ctx.Lookup(f.stopwordsTable, w); tw != nil {
			// замена на replacer-строку
			words[i] = tw.(string)
		}
	}
	return &types.Message{
		From:    m.From,
		To:      m.To,
		Content: strings.Join(words, " "),
	}
}

func (f *Filter) Run(ctx context.Context) func() error {
	return func() error {
		g := goka.DefineGroup(f.filterGroup,
			goka.Input(f.in, new(types.MessageCodec), f.filter),
			goka.Output(f.out, new(types.MessageCodec)),
			goka.Join(f.blockerTable, new(types.BlockValueCodec)),
			goka.Lookup(f.stopwordsTable, new(types.StopwordsValueCodec)),
		)
		p, err := goka.NewProcessor(f.brokers, g)
		if err != nil {
			return err
		}

		return p.Run(ctx)
	}
}
