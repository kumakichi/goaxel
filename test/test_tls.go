package main

import (
    "crypto/tls"
    //"crypto/x509"
    "fmt"
    "log"
)

func main() {
    /*
    cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
    if err != nil {
        log.Fatalf("server: loadkeys: %s", err)
    }
    config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
    */
    conn, err := tls.Dial("tcp", "github.com:443", nil)
    if err != nil {
        log.Fatalf("client: dial: %s", err)
    }
    defer conn.Close()
    log.Println("client: connected to: ", conn.RemoteAddr())

    /*
    state := conn.ConnectionState()
    for _, v := range state.PeerCertificates {
        fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
        fmt.Println(v.Subject)
    }
    log.Println("client: handshake: ", state.HandshakeComplete)
    log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)
    */
    _, err = conn.Write([]byte("GET /images/modules/home/gh-mac-app.png HTTP/1.0\r\nHost: github.com\r\nRang: bytes=1-\r\nUser-Agent: GoAxel 1.0\r\n\r\n"))
    if err != nil {
        log.Fatalf("client: write: %s", err)
    }

    data := make([]byte, 10240)
    n, err := conn.Read(data)
    fmt.Println(string(data[:n]))
    log.Print("client: exiting")
}
