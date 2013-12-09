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
    Desc    string
}

type ServerMet struct {
    Debug       bool
    IsValid     bool        `access:read`
    ServerCount int32       `access:read`
    Servers     []*Server   `access:read`
    buf         []byte
    offset      int
}

func (this *ServerMet) init() {
    this.offset = 0
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

func (this *ServerMet) parseTags(tagCount int32) (name, desc string) {
    name = ""
    desc = ""
    for i := 0; i < int(tagCount); i++ {
        tagType := this.buf[this.offset]
        if this.Debug {
            fmt.Printf("DEBUG: tagType 0x%02x\n", tagType)
        }
        if tagType == 0x02 {
            this.offset++
            numOfByteInSpecTag := this.byteToInt16(this.buf[this.offset:this.offset + 2])
            this.offset += 2
            if this.Debug {
                fmt.Println("DEBUG: number of bytes in Special tag", numOfByteInSpecTag)
            }
            specTagType := this.buf[this.offset:this.offset + int(numOfByteInSpecTag)]
            this.offset += int(numOfByteInSpecTag)
            if this.Debug {
                fmt.Println("DEBUG: Special tag type", specTagType)
            }
            strLen := this.byteToInt16(this.buf[this.offset:this.offset + 2])
            this.offset += 2
            strVal := string(this.buf[this.offset:this.offset + int(strLen)])
            if specTagType[0] == OP_LOGINREQUEST {
                name = strVal
            } else if specTagType[0] == 0x0b {
                desc = strVal
            }
            if this.Debug {
                fmt.Println("DEBUG: string length", strLen)
                fmt.Println("DEBUG:", strVal)
            }
            this.offset += int(strLen)
        } else if tagType == 0x03 {
            this.offset++
            numOfByteInSpecTag := this.byteToInt16(this.buf[this.offset:this.offset + 2])
            this.offset += 2
            if this.Debug {
                fmt.Println("DEBUG: number of bytes in Special tag", numOfByteInSpecTag)
            }
            specTagType := this.buf[this.offset:this.offset + int(numOfByteInSpecTag)]
            this.offset += int(numOfByteInSpecTag)
            if this.Debug {
                fmt.Println("Debug: Special tag type", specTagType)
                fmt.Println("DEBUG:", this.buf[this.offset:this.offset + 4])
            }
            this.offset += 4
        }
    }
    return
}

func (this *ServerMet) OpenFile(filePath string) {
    b, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
    }
    this.buf = b
    if this.buf[this.offset] == 0x0E || len(this.buf) < 6 {
        this.IsValid = true
        this.offset++
    } else {
        this.IsValid = false
        fmt.Println("ERROR: Invalid server met file")
        return
    }

    this.ServerCount = this.byteToInt32(this.buf[this.offset:5])
    this.offset += 4
    if this.Debug {
        fmt.Println("DEBUG: server count", this.ServerCount)
    }

    for i := 0; i < int(this.ServerCount); i++ {
        ipv4 := fmt.Sprintf("%v.%v.%v.%v",
            this.buf[this.offset], this.buf[this.offset + 1],
            this.buf[this.offset + 2], this.buf[this.offset + 3])
        this.offset += 4
        if this.Debug {
            fmt.Println("DEBUG: ipv4", ipv4)
        }
        port := this.byteToInt16(this.buf[this.offset:this.offset + 2])
        this.offset += 2
        if this.Debug {
            fmt.Println("DEBUG: port", port)
        }
        tagCount := this.byteToInt32(this.buf[this.offset:this.offset + 4])
        this.offset += 4
        if this.Debug {
            fmt.Println("DEBUG: tag count", tagCount)
        }
        name, desc := this.parseTags(tagCount)
        this.Servers = append(this.Servers,
            &Server{IP: ipv4, Port: port, Name: name, Desc: desc})
    }
}
