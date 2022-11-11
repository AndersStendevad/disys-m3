package main

import (
    "sync"
    "time"
    "math/rand"
    "fmt"
    "net"
    chat "github.com/AndersStendevad/disys-m3/grpc"
    "google.golang.org/grpc"
    "context"
)
type MessageEvent struct {
   Data interface{}
   Topic string
}

// DataChannel is a channel which can accept an MessageEvent
type DataChannel chan MessageEvent
// DataChannelSlice is a slice of DataChannels
type DataChannelSlice [] DataChannel

type EventBus struct {
   subscribers map[string]DataChannelSlice
   rm sync.RWMutex
}

func (eb *EventBus)Subscribe(topic string, ch DataChannel)  {
   eb.rm.Lock()
   if prev, found := eb.subscribers[topic]; found {
      eb.subscribers[topic] = append(prev, ch)
   } else {
      eb.subscribers[topic] = append([]DataChannel{}, ch)
   }
   eb.rm.Unlock()
}

func (eb *EventBus) Publish(topic string, data interface{}) {
   eb.rm.RLock()
   if chans, found := eb.subscribers[topic]; found {
      // this is done because the slices refer to same array even though they are passed by value
      // thus we are creating a new slice with our elements thus preserve locking correctly.
      channels := append(DataChannelSlice{}, chans...)
      go func(data MessageEvent, dataChannelSlices DataChannelSlice) {
         for _, ch := range dataChannelSlices {
            ch <- data
         }
      }(MessageEvent{Data: data, Topic: topic}, channels)
   }
   eb.rm.RUnlock()
}

var eb = &EventBus{
   subscribers: map[string]DataChannelSlice{},
}

func publisTo(topic string, data string)  {
   for {
      eb.Publish(topic, "Hi")
      eb.Publish(topic, data)
      eb.Publish(topic, "goodbye")
      time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
   }
}

func printMessageEvent(ch string, data MessageEvent)  {
   fmt.Printf("Channel: %s; Topic: %s; MessageEvent: %v\n", ch, data.Topic, data.Data)
}

type ChatServer struct {
    chat.UnimplementedChatServer
}

func main()  {
    lis, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Printf("failed to listen: %v", err)
    }
    var opts []grpc.ServerOption
    server := grpc.NewServer(opts...)
    chat.RegisterChatServer(server, &ChatServer{})


    if err := server.Serve(lis); err != nil {
        fmt.Printf("failed to serve: %v", err)
    }

}

func (s *ChatServer) Send(context.Context, *chat.Message) (*chat.MessageAck, error) {
    fmt.Println("Send")
    msg := chat.MessageAck{}
    return &msg, nil
}

func (s *ChatServer) Receive(*chat.Request, chat.Chat_ReceiveServer) error {
    fmt.Println("Receive")
    return nil
}

func old(){
    // old
    ch1 := make(chan MessageEvent)
    ch2 := make(chan MessageEvent)
    ch3 := make(chan MessageEvent)
    eb.Subscribe("topic1", ch1)
    eb.Subscribe("topic2", ch2)
    eb.Subscribe("topic2", ch3)
    go publisTo("topic1", "Hi topic 1")
    go publisTo("topic2", "Welcome to topic 2")
    go publisTo("topic3", "You joined topic 3")
    for {
        select {
        case d := <-ch1:
            go printMessageEvent("ch1", d)
        case d := <-ch2:
            go printMessageEvent("ch2", d)
        case d := <-ch3:
            go printMessageEvent("ch3", d)
        }
    }
}

func handle_connection(topic string) {
    ch := make(chan MessageEvent)
    eb.Subscribe(topic, ch)
    for {
        select {
        case d := <-ch:
            go printMessageEvent(topic, d)
        }
    }
}
// https://dev.bitolog.com/grpc-long-lived-streaming/
