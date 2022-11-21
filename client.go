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
    "fmt"
    "bufio"
    "os"
    "os/exec"
    "context"
    chat "github.com/AndersStendevad/disys-m3/grpc"
    "google.golang.org/grpc"
)


var input[]byte

func print(ctx context.Context, client chat.ChatClient, author string, topic string) {
    stream, err := client.Receive(ctx, &chat.Request{Author: author, Topic: topic})
    if err != nil {
        println("Error: %v", err)
    }
    defer stream.CloseSend()
    
    for {
        fmt.Printf("\r                                                        \r")
        message, err := stream.Recv()
        if err != nil {
            println("Error: %v", err)
            break
        }
        fmt.Println(message)
        fmt.Print(string(input))
   }
}

func main() {
    input = append(input, 0x3e)
    input = append(input, 0x3e)
    input = append(input, 0x3e)
    input = append(input, 0x20)
    
    author := os.Args[1]
    topic := os.Args[2]

    var opts []grpc.DialOption
    opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
    conn, err := grpc.Dial("localhost:8080", opts...)
    if err != nil {
        println("did not connect: %v", err)
    }
    println("Starting client")
    println("Joining as user:", author)
    println("To topic:", topic)
    println()
    fmt.Print(string(input))

    ctx := context.Background()
    client := chat.NewChatClient(conn)

    go print(ctx, client, author, topic)
    
    reader := bufio.NewReader(os.Stdin)
    exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
    for {
        b, err := reader.ReadByte()
        if err != nil {
                panic(err)
        }
        if b == 0x0A { // send on enter
            // send data
            fmt.Printf("\r                                                        \r")
            _, err := client.Send(ctx, &chat.Message{Author: author, Topic: topic, Message: string(input[4:])})
            if err != nil {
                println("Error: %v", err)
            }
            input = input[:4]
        } else {
        input = append(input, b)
        }
    }
}
