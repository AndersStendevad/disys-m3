package main

import (
    "sync"
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

func (s *ChatServer) Send(ctx context.Context, in *chat.Message) (*chat.MessageAck, error) {
    msg := in.Author+ ": " +in.Message
    fmt.Println("Server received message:", in)
    eb.Publish(in.Topic, msg)
    response := chat.MessageAck{Flag: "OK"}
    return &response, nil
}

func (s *ChatServer) Receive(msg *chat.Request, stream chat.Chat_ReceiveServer) error {
    ch := make(chan MessageEvent)
    eb.Subscribe(msg.Topic, ch)
    eb.Publish(msg.Topic, msg.Author + " joined")
    fmt.Println("Server received subscriber:", msg)
    for {
        select {
        case <-stream.Context().Done():
            eb.Publish(msg.Topic, msg.Author + " left")
            fmt.Println("Server lost subscriber:", msg)
            return nil
        case d := <-ch:
            stream.Send(&chat.Message{Message: d.Data.(string)})
        }
    }
    return nil
}

