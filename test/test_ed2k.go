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
    "bytes"
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

func byteToInt32(data []byte) (ret int32) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func int16ToByte(data int16) (ret []byte) {
    ret = []byte{}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.LittleEndian, data)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
        return
    }
    ret = buf.Bytes()
    return
}

func int32ToByte(data int32) (ret []byte) {
    ret = []byte{}
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.LittleEndian, data)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
        return
    }
    ret = buf.Bytes()
    return
}

func main() {
    fmt.Printf("0x%02x 0x%02x 0x%02x 0x%02x\n", 17, 60, 32, 251)
    fmt.Println(byteToInt32([]byte{29, 7, 0, 0}))
    fmt.Println(int16ToByte(int16(4662)))
    fmt.Println(int32ToByte(1821))

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
                   /*            string  numOfByte SpecialTag strLen   */
                   /* nickname */0x02,   1, 0,     1,         6, 0,  'l', 'e', 's', 'l', 'i', 'e',
                   /*            integer numOfByte SpecialTag intValue */
                   /* version  */0x03,   1, 0,     0x11,      0x3C, 0, 0, 0,
                   /* port tag */0x03,   1, 0,     0x0F,      29, 7, 0, 0,
                   /* flags    */0x03,   1, 0,     0x20,      128, 12, 4, 3}
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
