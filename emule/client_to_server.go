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
    "log"
    "net"
)

type Client2Server struct {
    Protocol    string
    Host        string
    Port        int
    ClientPort  int32
    NickName    string
    Debug       bool
    conn        net.Conn
}

func NewClient2Server(protocol, host string, port int, client_port int32,
                      nickname string, debug bool) *Client2Server {
    return &Client2Server{Protocol:   protocol,
                          Host:       host,
                          Port:       port,
                          ClientPort: client_port,
                          NickName:   nickname,
                          Debug:      debug}
}

func (this *Client2Server) Login() {
    u, _            := guid()
    client_port     := int16ToByte(int16(this.ClientPort))
    client_port32   := int32ToByte(this.ClientPort)
    lenNickName     := int16ToByte(int16(len(this.NickName)))
    buf             := []byte{OP_EDONKEYHEADER,
                              0, 0, 0, 0,
                              OP_LOGINREQUEST,
                              /* uuid */
                              u[0], u[1], u[2], u[3], u[4], u[5], u[6], u[7],
                              u[8], u[9], u[10], u[11], u[12], u[13], u[14],
                              u[15],
                              /* user id */
                              0, 0, 0, 0,
                              /* user port */
                              client_port[0], client_port[1],
                              4, 0, 0, 0,
                              /* tag nickname */
                              TAG_STRING, 1, 0, SPECIAL_TAG_NAME,
                              lenNickName[0], lenNickName[1]}
    /* append nickname */
    nickname        := []byte(this.NickName)
    for i := 0; i < len(nickname); i++ { buf = append(buf, nickname[i]) }
    /* append other tag */
    other_tag       := []byte{TAG_INTEGER, 1, 0, SPECIAL_TAG_VERSION,
                             0x3C, 0, 0, 0,
                             TAG_INTEGER, 1, 0, SPECIAL_TAG_PORT,
                             client_port32[0], client_port32[1],
                             client_port32[2], client_port32[3],
                             TAG_INTEGER, 1, 0, 0x20, 128, 12, 4, 3}
    for i := 0; i < len(other_tag); i++ { buf = append(buf, other_tag[i]) }
    lenBuf          := int32ToByte(int32(len(buf) - 5))
    for i := 1; i < 5; i++ { buf[i] = lenBuf[i - 1] }
    if this.Debug {
        fmt.Println("DEBUG:", buf)
        fmt.Println("DEBUG:", string(buf))
    }

    if this.conn != nil {
        _, err := this.conn.Write(buf)
        if err != nil { log.Println(err.Error()) }
    }
}

func (this *Client2Server) Connect() (ret bool) {
    ret = false
    conn, err := net.Dial(this.Protocol, fmt.Sprintf("%s:%d", this.Host, this.Port))
    if err != nil {
        this.conn = nil
        log.Println(err.Error())
        return
    }
    ret = true
    this.conn = conn
    return
}

func (this *Client2Server) Disconnect() {
    if this.conn != nil { defer this.conn.Close() }
}
