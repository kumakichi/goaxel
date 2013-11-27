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

func (h *HTTP) WriteToFile(range_from, range_to int, outputFileName string) {
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
    chunkName := fmt.Sprintf("%s.part.%d", outputFileName, range_from)
    f, err := os.Create(chunkName)
    if err != nil { panic(err) }
    defer f.Close()
    f.Write(data)
    if h.Callback != nil {
        h.Callback(len(data))
    }
}

func (h *HTTP) Get(url string) {
    h.url = fmt.Sprintf("http://%s:%d%s", h.host, h.port, url)
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
    ret, _ = strconv.Atoi(h.resp.Header.Get("Content-Length"))
    return ret
}
