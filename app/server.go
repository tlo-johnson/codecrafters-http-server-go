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

  requestComponents := strings.Split(string(request), "\r\n")
  path := strings.Split(requestComponents[0], " ")[1]
  switch {
    case path == "/":
      handleRootPath(connection)
      return

    case strings.HasPrefix(path, "/echo/"):
      handleEchoPath(connection, path)
      return

    case path == "/user-agent":
      handleUserAgentPath(connection, requestComponents)
      return

    default:
      handleUnknownPath(connection)
  }
}

func handleRootPath(connection net.Conn) {
  connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
  connection.Write([]byte("\r\n"))
}

func handleEchoPath(connection net.Conn, path string) {
  body, _ := strings.CutPrefix(path, "/echo/")
  sendResponse(connection, body)
}

func handleUnknownPath(connection net.Conn) {
  connection.Write([]byte("HTTP/1.1 404 Not Found\r\n"))
  connection.Write([]byte("\r\n"))
}

func handleUserAgentPath(connection net.Conn, lines []string) {
  for _, line := range lines {
    if !strings.HasPrefix(line, "User-Agent") {
      continue
    }

    userAgent, _ := strings.CutPrefix(line, "User-Agent: ")
    sendResponse(connection, userAgent)
  }
}

func sendResponse(connection net.Conn, body string) {
  connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
  connection.Write([]byte("Content-Type: text/plain\r\n"))
  connection.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
  connection.Write([]byte("\r\n"))
  connection.Write([]byte(body))
}

func exitOnError(err error, message string) {
  if err == nil {
    return
  }

  fmt.Printf("%s: %s", message, err.Error())
  os.Exit(1)
}
