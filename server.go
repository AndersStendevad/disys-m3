/*
┌──────────────────────────────┐  ┌──────────────────────────────┐
│                              │  │                              │
│  Client                      │  │  Server                      │
│  ------                      │  │  ------                      │
│                              │  │                              │
│  Can Request messages        │  │  Handles new Subscribers     │
│                              │  │                              │
│  Can Send messages to topic  │  │  Unsubscribes disconnected   │
│                              │  │  users                       │
│  Can use Lamport timstamp    │  │                              │
│  To figure out the order or  │  │  Increment lamport timestamp │
│  lost messages               │  │  on every action serverside  │
│                              │  │                              │
└──────────────────────────────┘  └──────────────────────────────┘
_________________________________________________________________________________

┌──────────────────────────────┐
│                              │
│  Client 1                    │
│                              │            ┌───────────────────────────────────┐
│  username : Anders           │ Receive    │                                   │
│  topic    : itu              ├───────────►│  Server                           │
│                              │   Messages │  ------                           │
│  Send     : Send 1 Message   │ ◄───────── │                                   │
│  Receive  : Get stream of    │   Send     │  EventBus struct:                 │
│             messages         ├───────────►│    Subscribe                      │
└──────────────────────────────┘            │    Unsubscribe                    │
                                            │    Publish                        │
┌──────────────────────────────┐            │                                   │
│                              │            │  Send gRPC:                       │
│  Client 2                    │            │    Will get a message and publish │
│                              │ Receive    │    it to the EventBus             │
│  username : Emil             ├───────────►│    Send back ACK to client        │
│  topic    : itu              │   Messages │                                   │
│                              │ ◄───────── │  Receive gRPC:                    │
│  Send     : Send 1 Message   │   Send     │    Open stream to client          │
│  Receive  : Get stream of    ├───────────►│                                   │
│             messages         │            │    Subscribe client to topic      │
└──────────────────────────────┘            │    through EventBus               │
                                            │                                   │
┌──────────────────────────────┐            │    Stream messages from EventBus  │
│                              │ Receive    │    to client                      │
│  Client 3                    ├───────────►│                                   │
│                              │   Messages │                                   │
│  username : Sebastian        │ ◄───────── │                                   │
│  topic    : itu              │   Send     │                                   │
│                              ├───────────►└───────────────────────────────────┘
│  Send     : Send 1 Message   │
│  Receive  : Get stream of    │
│             messages         │
└──────────────────────────────┘
 */

package main

import (
    "sync"
    "fmt"
    "net"
    chat "github.com/AndersStendevad/disys-m3/grpc"
    "google.golang.org/grpc"
    "context"
    "strconv"
)

type MessageEvent struct {
   Data interface{}
   Topic string
   lamport_timestamp int
}

type DataChannel chan MessageEvent

type DataChannelSlice [] DataChannel

type EventBus struct {
   subscribers map[string]DataChannelSlice
   rm sync.RWMutex
   lamport_timestamp int
}

func (eb *EventBus)Subscribe(topic string, ch DataChannel, msg *chat.Request) {
    eb.rm.Lock()
    eb.lamport_timestamp++
    fmt.Println("time:", eb.lamport_timestamp," Server received subscriber:", msg)

    if prev, found := eb.subscribers[topic]; found {
        eb.subscribers[topic] = append(prev, ch)
    } else {
        eb.subscribers[topic] = append([]DataChannel{}, ch)
    }
    eb.rm.Unlock()
}

func (eb *EventBus)Unsubscribe(topic string, ch DataChannel, msg *chat.Request) {
    eb.rm.Lock()
    eb.lamport_timestamp++
    fmt.Println("time:", eb.lamport_timestamp," Server lost subscriber:", msg)
    if prev, found := eb.subscribers[topic]; found {
        for i, c := range prev {
            if c == ch {
                eb.subscribers[topic] = append(prev[:i], prev[i+1:]...)
                break
            }
        }
    }
    eb.rm.Unlock()
}

func (eb *EventBus) Publish(topic string, data interface{}) {
    eb.rm.Lock()
    eb.lamport_timestamp++
    fmt.Println("time:", eb.lamport_timestamp," Server received message on topic:", topic, ", with message:", data)
    eb.lamport_timestamp++
    fmt.Println("time:", eb.lamport_timestamp," Server broadcast message to subscribers")
    if chans, found := eb.subscribers[topic]; found {
        channels := append(DataChannelSlice{}, chans...)
        go func(data MessageEvent, dataChannelSlices DataChannelSlice) {
            for _, ch := range dataChannelSlices {
                data.lamport_timestamp = eb.lamport_timestamp
                ch <- data
            }
        }(MessageEvent{Data: data, Topic: topic}, channels)
    }
    eb.rm.Unlock()
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
    eb.Publish(in.Topic, msg)
    response := chat.MessageAck{Flag: "OK"}
    return &response, nil
}

func (s *ChatServer) Receive(msg *chat.Request, stream chat.Chat_ReceiveServer) error {
    ch := make(chan MessageEvent)
    eb.Subscribe(msg.Topic, ch, msg)
    eb.Publish(msg.Topic, msg.Author + " joined")
    for {
        select {
        case <-stream.Context().Done():
            eb.Unsubscribe(msg.Topic, ch, msg)
            eb.Publish(msg.Topic, msg.Author + " left")
            return nil
        case d := <-ch:
            stream.Send(&chat.Message{Message: "Lamport timestamp: "+strconv.Itoa(d.lamport_timestamp) +" | "+ d.Data.(string)})
        }
    }
    return nil
}

