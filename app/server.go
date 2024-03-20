package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
  defer listener.Close()
	
  for {
    connection, err := listener.Accept()
    if err != nil {
      fmt.Println("Error accepting connection: ", err.Error())
      os.Exit(1)
    }
    go handleConnection(connection)
  }
}

func handleConnection(connection net.Conn) {
  defer connection.Close()

  connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
  connection.Write([]byte("\r\n"))
}
