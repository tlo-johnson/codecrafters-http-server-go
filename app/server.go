package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var directory string

func main() {
  parseFlags()

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
  exitOnError(err, "failed to bind to port 4221")

  defer listener.Close()
  for {
    connection, err := listener.Accept()
    exitOnError(err, "error accepting connection")

    go handleConnection(connection)
  }
}

func parseFlags() {
  flag.StringVar(&directory, "directory", "", "directory containing source files")
  flag.Parse()
}

func handleConnection(connection net.Conn) {
  defer connection.Close()

  request := make([]byte, 1024)
  _, err := connection.Read(request)
  exitOnError(err, "could not read request");

  var (
    method string
    path string
    body []byte
    contentLength int
  )
  requestContents := strings.Split(string(request), "\r\n")

  for index, value := range(requestContents) {
    if index == 0 {
      parts := strings.Split(value, " ")
      method = parts[0]
      path = parts[1]
    } else if strings.HasPrefix(value, "Content-Length") {
      length, _ := strings.CutPrefix(value, "Content-Length: ")
      contentLength, _ = strconv.Atoi(length)
    } else if value == "" {
      rawBody := strings.Join(requestContents[index + 1:], "\r\n")
      body = []byte(rawBody)
      body = body[:contentLength]
    }
  }

  switch {
    case path == "/":
      handleRootPath(connection)
      return

    case strings.HasPrefix(path, "/echo/"):
      handleEchoPath(connection, path)
      return

    case path == "/user-agent":
      handleUserAgentPath(connection, requestContents)
      return

    case method == "GET" && strings.HasPrefix(path, "/files/"):
      handleGetFilePath(connection, path)
      return

    case method == "POST" && strings.HasPrefix(path, "/files/"):
      handlePostFilePath(connection, path, body)
      return

    default:
      notFoundResponse(connection)
  }
}

func handleRootPath(connection net.Conn) {
  connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
  connection.Write([]byte("\r\n"))
}

func handleEchoPath(connection net.Conn, path string) {
  body, _ := strings.CutPrefix(path, "/echo/")
  sendResponse(connection, body, "text/plain")
}

func notFoundResponse(connection net.Conn) {
  connection.Write([]byte("HTTP/1.1 404 Not Found\r\n"))
  connection.Write([]byte("\r\n"))
}

func handleUserAgentPath(connection net.Conn, lines []string) {
  for _, line := range lines {
    if !strings.HasPrefix(line, "User-Agent") {
      continue
    }

    userAgent, _ := strings.CutPrefix(line, "User-Agent: ")
    sendResponse(connection, userAgent, "text/plain")
  }
}

func handleGetFilePath(connection net.Conn, path string) {
  relativePath, _ := strings.CutPrefix(path, "/files/")
  absolutePath := directory + "/" + relativePath
  contents, err := os.ReadFile(absolutePath)

  if err == nil {
    sendResponse(connection, string(contents), "application/octet-stream")
  } else {
    notFoundResponse(connection)
  }
}

func handlePostFilePath(connection net.Conn, path string, body []byte) {
  relativePath, _ := strings.CutPrefix(path, "/files/")
  absolutePath := directory + "/" + relativePath
  err := os.WriteFile(absolutePath, body, 0666)

  if err != nil {
    fmt.Println(err)
  }

  connection.Write([]byte("HTTP/1.1 201 OK\r\n\r\n"))
}

func sendResponse(connection net.Conn, body string, contentType string) {
  connection.Write([]byte("HTTP/1.1 200 OK\r\n"))
  connection.Write([]byte("Content-Type: " + contentType + "\r\n"))
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
