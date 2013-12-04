/*                                                                              
 * Copyright (C) 2013 Deepin, Inc.                                                 
 *               2013 Leslie Zhai <zhaixiang@linuxdeepin.com>                   
 *                                                                              
 * This program is free software: you can redistribute it and/or modify         
 * it under the terms of the GNU General Public License as published by         
 * the Free Software Foundation, either version 3 of the License, or            
 * any later version.                                                           
 *                                                                              
 * This program is distributed in the hope that it will be useful,              
 * but WITHOUT ANY WARRANTY; without even the implied warranty of               
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                
 * GNU General Public License for more details.                                 
 *                                                                              
 * You should have received a copy of the GNU General Public License            
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.        
 */

package main

import (
    "fmt"
    "net"
    "io"
    "os"
    "time"
    "encoding/binary"
)

func GUID() []byte {
    f, _ := os.Open("/dev/urandom")
    b := make([]byte, 16)
    f.Read(b)
    defer f.Close()
    uuid := fmt.Sprintf("%x-%x-%x-%x-%x-%x-%x-%x",
        b[0:2], b[2:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:14], b[14:16])
    fmt.Println("DEBUG:", uuid)
    return b
}

func le32(num int32) {
    b := make([]byte, 4)
    binary.LittleEndian.PutUint32(b, uint32(num))
    fmt.Println("DEBUG:", b)
}

func le16(num int32) {
    b := make([]byte, 2)
    binary.LittleEndian.PutUint16(b, uint16(num))
    fmt.Println("DEBUG:", b)
}

func respConn(conn net.Conn) {
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        println("ERROR:", err.Error())
        return
    }
    fmt.Println("DEBUG:", buf[:n])
    return
}

func main() {
    le32(4661)

    /* socket connect */
    //conn, err := net.Dial("tcp", "88.191.228.66:7111")
    conn, err := net.Dial("tcp", "0.0.0.0:7111")
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    } else {
        fmt.Println("Connected")
    }
    defer conn.Close()

    /* socket write */
    data := []byte{0xE3,
                   63, 0, 0, 0,
                   0x01,
                   0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                   0, 0, 0, 0,
                   54, 18,
                   4, 0, 0, 0,
                   2, 0x1, 0, 1, 6/* len(nickname) */, 0, 'l', 'e', 's', 'l', 'i', 'e',
                   /* version tag */3, 1, 0, 17, 60, 0, 0, 0,
                   /* port tag */3, 1, 0, 32, 29, 7, 0, 0,
                   /* flags tag */3, 1, 0, 251, 128, 12, 4, 3}
    uuid := GUID()
    for i := 6; i < 22; i++ { data[i] = uuid[i - 6] }
    fmt.Println("DEBUG:", string(data))
    fmt.Println("DEBUG:", data)
    fmt.Println("DEBUG:", len(data) - 5)
    _, err = conn.Write(data)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /*
    listener, err := net.Listen("tcp", "0.0.0.0:4662")
    if err != nil {
        println("error listening:", err.Error())
        os.Exit(1)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            println("Error accept:", err.Error())
            return
        }
        go respConn(conn)
    }
    */

    /* socket read */
    data = make([]byte, 1024)
    for {
        n, err := conn.Read(data)
        if err != nil {
            if err != io.EOF { panic(err) }
        }
        if n != 0 {
            fmt.Println(string(data[:n]))
        }
        time.Sleep(1000 * time.Millisecond)
    }
}
