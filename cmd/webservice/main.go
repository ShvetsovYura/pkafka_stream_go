package main

import (
	"context"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/webservice"
)

var (
	brokers = []string{"localhost:9092"}
)
var (
	blockerGroupName = "blocker"
	// filterGroupName    = "message_filter"
	collectorGroupName = "collector"
	stopwordsGroupName = "stopwords"

	blockerStreamName = "block_user"
	sentStreamName    = "message_sent"
	// receivedStreamName = "message_received"
	stopwordStreamName = "stopwords"
)

func main() {
	web := webservice.NewWebAPI(":8081",
		brokers,
		sentStreamName,
		blockerStreamName,
		blockerGroupName,
		collectorGroupName,
		stopwordStreamName,
		stopwordsGroupName,
	)
	web.Run(context.Background())
}
