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
    "flag"
    "os"
    "net/url"
    "strings"
    "strconv"
    "fmt"
    "github.com/xiangzhai/goaxel/conn"
)

const (
    app_name string = "GoAxel"
)

var (
    conn_num    uint
    user_agent  string
    urls        []string
)

func init() {
    flag.UintVar(&conn_num, "n", 3, "Specify maximum speed (bytes per second)")
    flag.StringVar(&user_agent, "U", app_name, "Set user agent")
}

func start_routine() {
    go func() {
        fmt.Println("DEBUG: start routine")
    }()
}

func parseUrl(strUrl string) {
    u, err := url.Parse(strUrl)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
        return
    }
    host := u.Host
    port := 80
    pos := strings.Index(host, ":")
    if pos != -1 {
        port, _ = strconv.Atoi(host[pos:len(host) - pos])
        host = host[0:pos]
    }
    conn := new(conn.CONN)
    conn.Protocol = u.Scheme
    conn.Host = host
    conn.Port = port
    conn.Path = u.Path 
    length := conn.GetContentLength()
    fmt.Println("DEBUG: content length", length)
}

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: goaxel [options] url1 [url2] [url...]")
        return
    }

    if os.Args[1] == "-V" {
        fmt.Println(fmt.Sprintf("%s Version 1.0", app_name))
        fmt.Println("Copyright 2013 Leslie Zhai")
        return
    }

    flag.Parse()

    for i := 1; i < len(os.Args); i++ {
        if !strings.HasPrefix(os.Args[i], "-") {
            urls = append(urls, os.Args[i])
        }
    }
    if len(urls) == 0 {
        fmt.Println("Invalid urls")
        return
    }
    if len(urls) == 1 {
        parseUrl(urls[0])
    }

    fmt.Println("num-connections:", conn_num)
}
