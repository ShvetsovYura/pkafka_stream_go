package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/blocker"
	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/collector"
	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/filter"
	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/stopwords"

	"golang.org/x/sync/errgroup"
)

var (
	brokers = []string{"localhost:9092"}

	blockerGroupName   = "blocker"
	filterGroupName    = "message_filter"
	collectorGroupName = "collector"
	stopwordsGroupName = "stopwords"

	blockerStreamName   = "block_user"
	sentStreamName      = "message_sent"
	receivedStreamName  = "message_received"
	stopwordsStreamName = "stopwords"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)

	log.Println("starting collector")
	c := collector.NewCollector(brokers, collectorGroupName, receivedStreamName)
	grp.Go(c.Run(ctx))

	log.Println("starting filter")
	f := filter.NewFilter(brokers, sentStreamName, receivedStreamName, filterGroupName, blockerGroupName, stopwordsGroupName)
	grp.Go(f.Run(ctx))

	log.Println("starting blocker")
	b := blocker.NewBlocker(brokers, blockerGroupName, blockerStreamName)
	grp.Go(b.Run(ctx))

	log.Println("starting stopwords")
	sw := stopwords.NewStopwords(brokers, stopwordsGroupName, stopwordsStreamName)
	grp.Go(sw.Run(ctx))

	waiter := make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-waiter:
	case <-ctx.Done():
	}
	cancel()
	if err := grp.Wait(); err != nil {
		log.Println(err)
	}
}
