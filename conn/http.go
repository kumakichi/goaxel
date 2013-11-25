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

package conn

import (
    "fmt"
    "net"
    "regexp"
    "strings"
    "strconv"
)

type HTTP struct {
    host    string
    port    int
    user    string
    passwd  string
    Debug   bool
    conn    net.Conn
    header  string
    content string
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
    n, err := http.conn.Read(data)
    if err != nil {
        defer http.conn.Close()
        http.Error = err
        fmt.Println("ERROR: ", err.Error())
        return
    }
    http.content = string(data[:n])
    http.conn.Close()
}

func (http *HTTP) Get(url string, range_from, range_to int) {
    http.AddHeader(fmt.Sprintf("GET %s HTTP/1.0", url))
    if range_to == 0 {
        http.AddHeader(fmt.Sprintf("Range: bytes=1-"))
    } else {
        http.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", range_from, range_to))
    }
    http.AddHeader("User-Agent: GoAxel 1.")
    _, http.Error = http.conn.Write([]byte(http.header))
    if http.Error != nil {
        fmt.Println("ERROR: ", http.Error.Error())
    }
}

func (http *HTTP) GetContentLength() int {
    ret := 0
    r, err := regexp.Compile(`Content-Length: (.*)`)
    if err != nil {
        http.Error = err
        fmt.Println("ERROR: ", err.Error())
        return ret
    }
    result := r.FindStringSubmatch(http.content)
    if len(result) != 0 {
        s := strings.TrimSuffix(result[1], "\r")
        ret, http.Error = strconv.Atoi(s)
    }
    return ret
}
