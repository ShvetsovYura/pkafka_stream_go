package stopwords

import (
	"context"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/types"
	"github.com/lovoo/goka"
)

type Stopwords struct {
	brokers         []string
	stopwordsGroup  goka.Group
	stopwordsStream goka.Stream
}

func NewStopwords(brokers []string, stopwordsGroupName string, stopwordsStreamName string) *Stopwords {
	stopwordsGroup := goka.Group(stopwordsGroupName)
	stopwordsStream := goka.Stream(stopwordsStreamName)
	return &Stopwords{
		brokers:         brokers,
		stopwordsGroup:  stopwordsGroup,
		stopwordsStream: stopwordsStream,
	}
}
func setWords(ctx goka.Context, msg interface{}) {
	ctx.SetValue(msg.(string))
}

func (s *Stopwords) Run(ctx context.Context) func() error {
	return func() error {
		g := goka.DefineGroup(s.stopwordsGroup,
			goka.Input(s.stopwordsStream, new(types.StopwordsValueCodec), setWords),
			goka.Persist(new(types.StopwordsValueCodec)),
		)
		p, err := goka.NewProcessor(s.brokers, g)
		if err != nil {
			return err
		}

		return p.Run(ctx)
	}
}
