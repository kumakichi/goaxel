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

package emule

import (
    "fmt"
    "io/ioutil"
    "bytes"
    "encoding/binary"
)

type Server struct {
    IP      string
    Port    int16
    Name    string
}

type ServerMet struct {
    Debug       bool
    IsValid     bool        `access:read`
    ServerCount int32       `access:read`
    Servers     []Server    `access:read`
    buf         []byte
}

func (this *ServerMet) byteToInt16(data []byte) (ret int16) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func (this *ServerMet) byteToInt32(data []byte) (ret int32) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func (this *ServerMet) parseTags(tagCount int32, offset int) {
    tagType := this.buf[offset]
    if this.Debug {
        fmt.Printf("DEBUG: tagType 0x%02x\n", tagType)
    }
    if tagType == 0x02 {
        offset++
        numOfByteInSpecTag := this.byteToInt16(this.buf[offset:offset + 2])
        if this.Debug {
            fmt.Println("DEBUG:", numOfByteInSpecTag)
        }
    }
}

func (this *ServerMet) OpenFile(filePath string) {
    b, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
    }
    this.buf = b
    if this.buf[0] == 0x0E || len(this.buf) < 6 {
        this.IsValid = true
    } else {
        this.IsValid = false
        fmt.Println("ERROR: Invalid server met file")
        return
    }

    this.ServerCount = this.byteToInt32(this.buf[1:5])
    if this.Debug {
        fmt.Println("DEBUG:", this.ServerCount)
    }

    for i := 0; i < int(this.ServerCount); i++ {
        ipv4 := fmt.Sprintf("%v.%v.%v.%v",
            this.buf[i + 5], this.buf[i + 6], this.buf[i + 7], this.buf[i + 8])
        if this.Debug {
            fmt.Println("DEBUG:", ipv4)
        }
        port := this.byteToInt16(this.buf[i + 9:i + 11])
        if this.Debug {
            fmt.Println("DEBUG:", port)
        }
        tagCount := this.byteToInt32(this.buf[i + 11:i + 15])
        if this.Debug {
            fmt.Println("DEBUG:", tagCount)
        }
        this.parseTags(tagCount, i + 15)
        //break
    }
}
