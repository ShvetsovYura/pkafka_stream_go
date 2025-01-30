package webservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/collector"
	"github.com/ShvetsovYura/pkafka_stream_go.git/internal/types"
	"github.com/gorilla/mux"
	"github.com/lovoo/goka"
)

type WebApiService struct {
	address         string
	brokers         []string
	messageEmitter  *goka.Emitter
	blockerEmitter  *goka.Emitter
	stopwordEmitter *goka.Emitter
	messageView     *goka.View
	blockerView     *goka.View
	stopwordsView   *goka.View
	messageStream   goka.Stream
	blockerStream   goka.Stream
	stopwordsStream goka.Stream
}

func NewWebAPI(
	address string,
	brokers []string,
	messageStream string,
	blockerStream string,
	blockerGroupName string,
	collectorGroupName string,
	stopwordsStreamName string,
	stopwordsGroupName string,
) *WebApiService {
	collectorGroup := goka.Group(collectorGroupName)
	messageView, err := goka.NewView(brokers, goka.GroupTable(collectorGroup), new(collector.MessageListCodec))
	if err != nil {
		panic(err)
	}
	blockerGroup := goka.Group(blockerGroupName)
	blockerView, err := goka.NewView(brokers, goka.GroupTable(blockerGroup), new(collector.MessageListCodec))
	if err != nil {
		panic(err)
	}
	stopwordsGroup := goka.Group(stopwordsGroupName)
	stopwordsTable := goka.GroupTable(stopwordsGroup)
	stopwordsView, err := goka.NewView(brokers, stopwordsTable, new(types.StopwordsValueCodec))
	if err != nil {
		panic(err)
	}
	return &WebApiService{
		address:         address,
		brokers:         brokers,
		messageView:     messageView,
		blockerView:     blockerView,
		stopwordsView:   stopwordsView,
		messageStream:   goka.Stream(messageStream),
		blockerStream:   goka.Stream(blockerStream),
		stopwordsStream: goka.Stream(stopwordsStreamName),
	}
}

func (s *WebApiService) Run(ctx context.Context) {
	go s.messageView.Run(ctx)
	go s.blockerView.Run(ctx)
	go s.stopwordsView.Run(ctx)

	// вынести запуск в отдельное место и сделать отправку через каналы
	messageEmmiter, err := goka.NewEmitter(s.brokers, s.messageStream, new(types.MessageCodec))
	if err != nil {
		panic(err)
	}
	defer messageEmmiter.Finish()

	blockEmitter, err := goka.NewEmitter(s.brokers, s.blockerStream, new(types.BlockEventCodec))
	if err != nil {
		panic(err)
	}
	defer blockEmitter.Finish()

	stopwordsEmmiter, err := goka.NewEmitter(s.brokers, s.stopwordsStream, new(types.StopwordsValueCodec))
	if err != nil {
		panic(err)
	}
	defer stopwordsEmmiter.Finish()

	s.messageEmitter = messageEmmiter
	s.blockerEmitter = blockEmitter
	s.stopwordEmitter = stopwordsEmmiter

	router := mux.NewRouter()
	router.HandleFunc("/{user}/send", s.sendMessage()).Methods("POST")
	router.HandleFunc("/{user}/block", s.blockUser()).Methods("POST")
	router.HandleFunc("/{user}/who_blocked", s.getBlockedBy()).Methods("GET")
	router.HandleFunc("/{user}/messages", s.userIncomingMessages()).Methods("GET")
	router.HandleFunc("/stopwords/add", s.addStopword()).Methods("POST")
	router.HandleFunc("/stopwords/remove", s.removeStopword()).Methods("POST")

	log.Printf("Listen port 8081")
	log.Fatal(http.ListenAndServe(s.address, router))
}

func (s *WebApiService) sendMessage() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m types.Message

		b, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		m.From = mux.Vars(r)["user"]

		err = s.messageEmitter.EmitSync(m.From, &m)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		log.Printf("Sent message:\n %v\n", m)
		fmt.Fprintf(w, "Sent message:\n %v\n", m)
	}
}

func (s *WebApiService) userIncomingMessages() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		val, _ := s.messageView.Get(user)
		if val == nil {
			fmt.Fprintf(w, "%s not found!", user)
			return
		}
		messages := val.([]types.Message)
		fmt.Fprintf(w, "Latest messages for %s\n", user)
		for i, m := range messages {
			fmt.Fprintf(w, "%d %10s: %v\n", i, m.From, m.Content)
		}
	}
}

func (s *WebApiService) blockUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m types.BlockEvent

		b, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		user := mux.Vars(r)["user"]

		err = s.blockerEmitter.EmitSync(m.User, &types.BlockEvent{User: user, IsBlocked: m.IsBlocked})
		if err != nil {
			panic(err)
		}
		log.Printf("Sent block user:\n %v\n", m)
		fmt.Fprintf(w, "Sent block user:\n %v\n", m)
	}
}

func (s *WebApiService) getBlockedBy() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mux.Vars(r)["user"]
		val, _ := s.blockerView.Get(user)
		if val == nil {
			fmt.Fprintf(w, "%s not found!", user)
			return
		}
		blockers := val.(types.BlockValue)
		fmt.Fprintf(w, "Users who blocked %s\n", user)
		for _, u := range blockers.BlockedByUsers {
			fmt.Fprintf(w, " %10s: %s\n", u)
		}
	}
}

func (s *WebApiService) addStopword() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var sw types.StopkwordInput

		b, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		err = json.Unmarshal(b, &sw)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		err = s.stopwordEmitter.EmitSync(sw.Key, sw.Replacer)
		if err != nil {
			fmt.Printf("error on emit stopword %s\n", err.Error())
			return
		}
		log.Printf("Sent stopwords: %s\n", string(b))
	}
}

func (s *WebApiService) removeStopword() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var sw types.StopkwordRemoveInput

		b, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		err = json.Unmarshal(b, &sw)
		if err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		err = s.stopwordEmitter.EmitSync(sw.Key, "")
		if err != nil {
			fmt.Printf("error on emit stopword %s\n", err.Error())
			return
		}
		log.Printf("Sent stopwords: %s\n", string(b))
	}
}
