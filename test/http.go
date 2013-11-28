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
    "net/http"
    "strconv"
    "os"
    "io"
)

type HTTP struct {
    Protocol    string
    host        string
    port        int
    user        string
    passwd      string
    Debug       bool
    UserAgent   string
    resp        *http.Response
    url         string
    rangeFrom   int
    rangeTo     int
    offset      int
    Error       error
    Callback    func(int)
}

const (
    buffer_size int64 = 10240   // but 102400 io.LimitReader will be blocked
)

func (h *HTTP) Connect(host string, port int) {
    h.host = host
    h.port = port
}

func (h *HTTP) Response() {}

func (h *HTTP) WriteToFile(f *os.File) {
    client := &http.Client{}
    req, _ := http.NewRequest("GET", h.url, nil)
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", h.rangeFrom, h.rangeTo))
    req.Header.Set("User-Agent", h.UserAgent)
    if h.Debug {
        fmt.Println("DEBUG:", req.Header)
    }
    resp, err := client.Do(req)
    if err == io.EOF { return }
    if err != nil { panic(err) }
    defer resp.Body.Close()
    for {
        lr := io.LimitReader(resp.Body, buffer_size)
        data := make([]byte, buffer_size)
        n, err := io.ReadAtLeast(lr, data, int(buffer_size))
        if err != nil { return }
        f.WriteAt(data, int64(h.offset))
        if h.Callback != nil {
            h.Callback(n)
        }
        h.offset += n
    }
}

func (h *HTTP) Get(url string, range_from, range_to int) {
    h.rangeFrom = range_from
    h.rangeTo = range_to
    h.offset = h.rangeFrom
    if h.Protocol == "https" && h.port == 80 {
        h.port = 443
    }
    h.url = fmt.Sprintf("%s://%s:%d%s", h.Protocol, h.host, h.port, url)
    h.resp, h.Error = http.Head(h.url)
    if h.Error != nil {
        fmt.Println("ERROR:", h.Error.Error())
    }
}

func (h *HTTP) IsAcceptRange() bool {
    ret := true
    if h.resp.Header.Get("Accept-Ranges") == "" {
        ret = false
    }
    return ret
}

func (h *HTTP) GetContentLength() int {
    ret := 0
    ret, h.Error = strconv.Atoi(h.resp.Header.Get("Content-Length"))
    return ret
}
