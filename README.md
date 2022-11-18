# Mandatory 3
### By Anders and Emil

## Report
This is a server-client architecture using streaming from the server to each client. 
To make the clinetside code easy to work with, we decide to keep as much logic serverside as possible. Therefore we only implement 2 gRPC methods:

**Send**
To send a message to the server. This method does not require that a receive stream is running, but this could be implemented serverside.
**Receive**
From the clients perspective it is just a request to get a stream of messages on a topic. Subscribe and UnSubscribe are handled serverside.

Here is the proto file:
We use MessageAck as a flag to acknowledge that a published message went trough. Other than that Messages need author, topic and message. Requests only need author and topic. The Message message is reused for Send and Receive.

```
service Chat {
    rpc Send (Message) returns (MessageAck) {}
    rpc Receive (Request) returns (stream Message) {}
}

message Message {
    string author = 1;
    string topic = 2;
    string message = 3;
}

message MessageAck {
    string flag = 1;
}

message Request {
    string author = 1;
    string topic = 2;
}```



## Running the code

Starting the server by running this command. The server takes no arguments
<code>go run server.go</code>

You can start chatting on a topic with the follwing code. You have to provide a name and a topic. Multiple topics may be live at the same time and will increament the Lamport timestamp. 

<code>go run client.go NAME TOPIC</code>

for example:

<code>go run client.go Anders itu</code>
<code>go run client.go Emil itu</code>
<code>go run client.go Sebastian itu</code>

Each client will wait for the server to be live

## Client
The client has two go routines. One that sends messages and one the prints the incoming broadcasts. These run at the same time. You send a message by writing in the terminal and pressing \<ENTER\>.

You can disconnect with \<ctrl + c\>.

## Server
The server works concurrently and has as many connections open as clients. These have a bi-directional stream open to be able to send messages back to the clients when they come in.

When a users joins, the server will publish a message to the chat. If the connection is dropped, the server will publish a message with user left before closing the go routine.

The server has no log of messages. Instead an Eventbus is used. This is a struct which contain channels to connected users. When an event is published to the eventbus on a topic, all channels (clients connected) will get a copy of the messages.

## Example of running code:

```
    go run server.go 
    time: 1  Server received subscriber: author:"Anders"  topic:"itu"
    time: 2  Server received message on topic: itu , with message: Anders joined
    time: 3  Server broadcast message to subscribers
    time: 4  Server received subscriber: author:"Emil"  topic:"itu"
    time: 5  Server received message on topic: itu , with message: Emil joined
    time: 6  Server broadcast message to subscribers
    time: 7  Server received message on topic: itu , with message: Anders: Hello Emil
    time: 8  Server broadcast message to subscribers
    time: 9  Server received message on topic: itu , with message: Emil: Hello Anders
    time: 10  Server broadcast message to subscribers
    time: 11  Server received subscriber: author:"Sebastian"  topic:"itu"
    time: 12  Server received message on topic: itu , with message: Sebastian joined
    time: 13  Server broadcast message to subscribers
    time: 14  Server lost subscriber: author:"Sebastian"  topic:"itu"
    time: 15  Server received message on topic: itu , with message: Sebastian left
    time: 16  Server broadcast message to subscribers
    time: 17  Server lost subscriber: author:"Emil"  topic:"itu"
    time: 18  Server received message on topic: itu , with message: Emil left
    time: 19  Server broadcast message to subscribers
    time: 20  Server received message on topic: itu , with message: Anders: Now I am all alone :(
    time: 21  Server broadcast message to subscribers
    time: 22  Server received subscriber: author:"Emil"  topic:"itu"
    time: 23  Server received message on topic: itu , with message: Emil joined
    time: 24  Server broadcast message to subscribers
    time: 25  Server received message on topic: itu , with message: Emil: sorry connection issues
    time: 26  Server broadcast message to subscribers
```

```
    go run client.go Anders itu
    Starting client
    Joining as user: Anders
    To topic: itu

    message:"Lamport timestamp: 3 | Anders joined"          
    message:"Lamport timestamp: 6 | Emil joined"            
    Hello Emil                                              
    message:"Lamport timestamp: 8 | Anders: Hello Emil"     
    message:"Lamport timestamp: 10 | Emil: Hello Anders"    
    message:"Lamport timestamp: 13 | Sebastian joined"      
    message:"Lamport timestamp: 16 | Sebastian left"        
    message:"Lamport timestamp: 19 | Emil left"             
    Now I am all alone :(                                   
    message:"Lamport timestamp: 21 | Anders: Now I am all alone :("
    message:"Lamport timestamp: 24 | Emil joined"           
    message:"Lamport timestamp: 26 | Emil: sorry connection issues"
```

```
    go run client.go Emil itu
    Starting client
    Joining as user: Emil
    To topic: itu

    message:"Lamport timestamp: 6 | Emil joined"            
    message:"Lamport timestamp: 8 | Anders: Hello Emil"     
    Hello Anders                                            
    message:"Lamport timestamp: 10 | Emil: Hello Anders"    
    message:"Lamport timestamp: 13 | Sebastian joined"      
    message:"Lamport timestamp: 16 | Sebastian left"        
    Who t^?a^Csignal: interrupt 

    $ go run client.go Emil itu
    Starting client
    Joining as user: Emil
    To topic: itu

    message:"Lamport timestamp: 24 | Emil joined"           
    sorry connection issues                                 
    message:"Lamport timestamp: 26 | Emil: sorry connection issues"
```
```
    go run client.go Sebastian itu
    Starting client
    Joining as user: Sebastian
    To topic: itu

    message:"Lamport timestamp: 13 | Sebastian joined"      
    What are you talking b^?abo^?^Csignal: interrupt
```
