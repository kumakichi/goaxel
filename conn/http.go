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
    "io/ioutil"
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
    offset      int
    Error       error
    Callback    func(int)
}

func (h *HTTP) Connect(host string, port int) {
    h.host = host
    h.port = port
}

func (h *HTTP) Response() {}

func (h *HTTP) WriteToFile(range_from, range_to int, f *os.File) {
    client := &http.Client{}
    req, _ := http.NewRequest("GET", h.url, nil)
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", range_from, range_to))
    if h.Debug {
        fmt.Println("DEBUG:", req.Header)
    }
    resp, err := client.Do(req)
    if err == io.EOF { return }
    if err != nil { panic(err) }
    defer resp.Body.Close()
    data, err := ioutil.ReadAll(resp.Body)
    /* TODO: it does not need to write to tmp chunk file */
    f.WriteAt(data, int64(range_from))
    if h.Callback != nil {
        h.Callback(len(data))
    }
}

func (h *HTTP) Get(url string) {
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
    if h.Error != nil {
        fmt.Println("ERROR:", h.Error.Error())
    }
    return ret
}
