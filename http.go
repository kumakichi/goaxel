package http

import (
    "fmt"
    "net"
)

type HTTP struct {
    host    string
    port    int
    user    string
    passwd  string
    conn    net.Conn
    header  string
    Error   error
}

const (
    buffer_size int = 1024
)

func (http *HTTP) Connect(host string, port int) {
    http.conn, http.Error = net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
    if http.Error != nil {
        fmt.Println("ERROR: ", http.Error.Error())
        return
    }
    http.host = host
    http.port = port
}

func (http *HTTP) AddHeader(header string) {
    http.header += header + "\r\n\r\n"
}

func (http *HTTP) Response() {
    data := make([]byte, buffer_size)
    n, http.Error = http.conn.Read(data)
    if http.Error != nil {
        fmt.Println("ERROR: ", http.Error.Error())
        defer http.conn.Close()
        return
    }
    content := string(data[:n])
    fmt.Println("DEBUG: ", content)
    http.conn.Close()
}

func (http *HTTP) Get(url string) {
    http.AddHeader("GET " + url + " HTTP/1.0")
    http.AddHeader("Range: bytes=1-")
    http.AddHeader("User-Agent: GoAxel 1.")
    _, http.Error = http.conn.Write([]byte(http.header))
    if http.Error != nil {
        fmt.Println("ERROR: ", http.Error.Error())
    }
}
