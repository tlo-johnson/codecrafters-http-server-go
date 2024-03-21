package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
  exitOnError(err, "failed to bind to port 4221")

  defer listener.Close()
  for {
    connection, err := listener.Accept()
    exitOnError(err, "error accepting connection")

    go handleConnection(connection)
  }
}

func handleConnection(connection net.Conn) {
  defer connection.Close()

  request := make([]byte, 1024)
  _, err := connection.Read(request)
  exitOnError(err, "could not read request");

  path := strings.Split(string(request), " ")[1]
  if strings.HasPrefix(path, "/echo/") {
    body, _ := strings.CutPrefix(path, "/echo/")

    connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
    connection.Write([]byte("Content-Type: text/plain\r\n"))
    connection.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
    connection.Write([]byte("\r\n"))
    connection.Write([]byte(body))
  } else {
    connection.Write([]byte("HTTP/1.1 404 Not Found\r\n"))
    connection.Write([]byte("\r\n"))
  }
}

func exitOnError(err error, message string) {
  if err == nil {
    return
  }

  fmt.Printf("%s: %s", message, err.Error())
  os.Exit(1)
}
