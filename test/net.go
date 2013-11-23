package main

import (
    "fmt"
    "net"
    "regexp"
)

const (
    BUFFER_SIZE int = 1024
)

func main() {
    /* socket connect */
    conn, err := net.Dial("tcp", "localhost:80")
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /* socket write */
    _, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\nRange: bytes=1-\r\n\r\nUser-Agent: GoAxel 1.0\r\n\r\n"))
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /* socket read */
    data := make([]byte, BUFFER_SIZE)
    _, err = conn.Read(data)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        conn.Close()
        return;
    }
    s := string(data[:BUFFER_SIZE])
    fmt.Println("DEBUG: ", s)
    conn.Close()

    /* parse http header */
    r, err := regexp.Compile(`Content-Length: (.*)`)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
    }
    result := r.FindStringSubmatch(s)
    if len(result) != 0 {
        fmt.Println("DEBUG: content length ", result[1])
    }
}
