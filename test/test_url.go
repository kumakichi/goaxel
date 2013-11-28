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
    "net/url"
    "strings"
    "path"
)

func main() {
    u, err := url.Parse("http://localhost/leslie/movie/show/test.mp4")
    if err != nil {
        fmt.Println("ERROR:", err.Error())
    }
    fmt.Println("protocol:", u.Scheme)
    fmt.Println("host:", u.Host)
    if !strings.Contains(u.Host, ":") {
        fmt.Println("port: 80")
    }
    fmt.Println("path:", u.Path)
    fmt.Println("base:", path.Base(u.Path))

    fmt.Println("")

    u, err = url.Parse("ftp://anonymous:123456@localhost/1/a/test.png")
    //u, err = url.Parse("ftp://localhost/1/test.png")
    if err != nil {
        fmt.Println("ERROR:", err.Error())
    }
    fmt.Println("protocol:", u.Scheme)
    fmt.Println("host:", u.Host)
    if !strings.Contains(u.Host, ":") {
        fmt.Println("port: 21")
    }
    userinfo := u.User
    username := ""
    passwd := ""
    if userinfo != nil {
        username = userinfo.Username()
        passwd, _ = userinfo.Password()
    }
    fmt.Println("userinfo:", username, passwd)
    fmt.Println("path:", u.Path)
    fmt.Println("dir:", path.Dir(u.Path))
    fmt.Println("base:", path.Base(u.Path))
}
