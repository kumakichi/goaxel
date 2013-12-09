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
    "os"
    "io"
)

type HTTP struct {
    Protocol        string
    host            string
    port            int
    user            string
    passwd          string
    Debug           bool
    UserAgent       string
    conn            net.Conn
    header          string
    headerResponse  string
    offset          int
    Error           error
    Callback        func(int)
}

const (
    buffer_size int = 102400
)

func (this *HTTP) Connect(host string, port int) bool {
    address := fmt.Sprintf("%s:%d", host, port)
    this.conn, this.Error = net.Dial("tcp", address)
    if this.Error != nil {
        fmt.Println("ERROR: ", this.Error.Error())
        return false
    }
    this.host = host
    this.port = port
    return true
}

func (this *HTTP) AddHeader(header string) {
    this.header += header + "\r\n"
}

/* TODO: get this header */
func (this *HTTP) Response() {
    defer this.conn.Close()
    data := make([]byte, 1)
    for i := 0; ; {
        n, err := this.conn.Read(data)
        if err != nil {
            if err != io.EOF {
                this.Error = err
                fmt.Println("ERROR:", this.Error.Error())
                return
            }
        }
        if data[0] == '\r' {
            continue
        } else if data[0] == '\n' {
            if i == 0 {
                break
            }
            i = 0
        } else {
            i++
        }
        this.headerResponse += string(data[:n])
    }
    if this.Debug {
        fmt.Println("DEBUG:", this.headerResponse)
    }
    this.conn.Close()
}

func (this *HTTP) WriteToFile(outputFileName string, old_range_from int, chunkSize int) {
    this.offset = chunkSize
    defer this.conn.Close()
    resp := ""
    data := make([]byte, 1)
    for i := 0; ; {
        data := make([]byte, 1)
        n, err := this.conn.Read(data)
        if err != nil {
            if err != io.EOF {
                this.Error = err
                fmt.Println("ERROR:", this.Error.Error())
                return
            }
        }
        if data[0] == '\r' {
            continue
        } else if data[0] == '\n' {
            if i == 0 {
                break
            }
            i = 0
        } else {
            i++
        }
        resp += string(data[:n])
    }
    if this.Debug {
        fmt.Println("DEBUG:", resp)
    }
    chunkName := fmt.Sprintf("%s.part.%d", outputFileName, old_range_from)
    f, err := os.OpenFile(chunkName, os.O_CREATE | os.O_WRONLY, 0664)
    defer f.Close()
    if err != nil { panic(err) }
    data = make([]byte, buffer_size)
    for {
        n, err := this.conn.Read(data)
        if err != nil {
            if err != io.EOF {
                this.Error = err
                fmt.Println("ERROR:", this.Error.Error())
                return
            }
        }
        f.WriteAt(data[:n], int64(this.offset))
        if this.Callback != nil {
            this.Callback(n)
        }
        this.offset += n
        if err == io.EOF { return }
    }
    return
}

func (this *HTTP) Get(url string, range_from, range_to int) {
    this.AddHeader(fmt.Sprintf("GET %s HTTP/1.0", url))
    this.AddHeader(fmt.Sprintf("Host: %s", this.host))
    if range_to == 0 {
        this.AddHeader(fmt.Sprintf("Range: bytes=0-"))
    } else {
        this.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", range_from, range_to))
    }
    this.AddHeader(fmt.Sprintf("User-Agent: %s", this.UserAgent))
    this.AddHeader("")
    if this.Debug {
        fmt.Println("DEBUG:", this.header)
    }
    _, this.Error = this.conn.Write([]byte(this.header))
    if this.Error != nil {
        fmt.Println("ERROR: ", this.Error.Error())
    }
}

func (this *HTTP) IsAcceptRange() bool {
    ret := false

    if strings.Contains(this.headerResponse, "Content-Range") || 
        strings.Contains(this.headerResponse, "Accept-Ranges"){
        ret = true
    }

    return ret
}

func (this *HTTP) GetContentLength() int {
    ret := 0
    r, err := regexp.Compile(`Content-Length: (.*)`)
    if err != nil {
        this.Error = err
        fmt.Println("ERROR: ", err.Error())
        return ret
    }
    result := r.FindStringSubmatch(this.headerResponse)
    if len(result) != 0 {
        s := strings.TrimSuffix(result[1], "\r")
        ret, this.Error = strconv.Atoi(s)
    }
    return ret
}
