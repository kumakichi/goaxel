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
    "encoding/binary"
)

func uuid() []byte {
    f, _ := os.Open("/dev/urandom")
    b := make([]byte, 16)
    f.Read(b)
    defer f.Close()
    uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
    fmt.Println("DEBUG:", uuid)
    return b
}

func main() {
    /* socket connect */
    conn, err := net.Dial("tcp", "88.191.228.66:7111")
    defer conn.Close()
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    var i int32 = 64
    b := make([]byte, 4)
    binary.LittleEndian.PutUint32(b, uint32(i))
    fmt.Println("DEBUG:", b)

    /* socket write */
    data := []byte{0xe3,
                   64, 0, 0, 0,
                   0x01,
                   0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                   0, 0, 0, 0,
                   54, 18,
                   4, 0, 0, 0,
                   2, 0x1, 'l', 'e', 's', 'l', 'i', 'e',
                   3, 0x11, 60, 0, 0, 0, 0, 0, 0, 0,
                   3, 0x0f, 54, 18, 0, 0, 0, 0, 0, 0,
                   3, 0x20, 1, 0, 0, 0, 0, 0, 0, 0}
    tmp := uuid()
    for i := 6; i < 22; i++ { data[i] = tmp[i - 6] }
    fmt.Println("DEBUG:", data)
    fmt.Println("DEBUG:", len(data) - 6)
    _, err = conn.Write(data)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /* socket read */
    data = make([]byte, 1)
    resp := ""
    for {
        n, err := conn.Read(data)
        if err != nil {
            if err != io.EOF { panic(err) }
        }
        resp += string(data[:n])
        if err == io.EOF { return }
    }
    fmt.Println("DEBUG:", resp)
}
