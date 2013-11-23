package main

import (
    "fmt"
    "net"
)

const (
    BUFFER_SIZE int = 1024
)

func main() {
    var host string = "localhost";

    conn, err := net.Dial("tcp", host + ":80")
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
    }

    /* socket write */
    _, err = conn.Write([]byte("GET /test.jpg HTTP/1.0\r\n\r\nHost: localhost\r\n\r\nRange: bytes=1-\r\n\r\nUser-Agent: GoAxel 1.0\r\n\r\n"))

    /* socket read */
    for {
        data := make([]byte, BUFFER_SIZE)
        _, err := conn.Read(data)
        s := string(data[:BUFFER_SIZE])
        fmt.Println(s)
        if err != nil {
            fmt.Println("DEBUG: read EOF")
            conn.Close()
            break
        }
    }
}
