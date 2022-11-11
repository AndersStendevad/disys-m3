package main

import (
    "time"
    "math/rand"
    "fmt"
    "bufio"
    "os"
    "os/exec"
)

type client struct {


}

var input[]byte

func print() {
    var i int =0
    for {
      i++
      // fetch data to print here
      fmt.Printf("\r                                                        \r")

      // fetch from server
      message := "Data from server: Hello World"

      // print
      fmt.Println(message)
      fmt.Print(string(input))
      time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
   }
}

func main() {
    input = append(input, 0x3e)
    input = append(input, 0x3e)
    input = append(input, 0x3e)
    input = append(input, 0x20)
    go print()
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
            fmt.Print("Data sent to server: ", string(input[4:]),"\n>>> ")
            input = input[:4]
        } else {
        input = append(input, b)
        }
    }
}
