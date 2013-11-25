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
)

type CONN struct {
    Protocol    string
    Host        string
    Port        int
    User        string
    Passwd      string
    Path        string
    Debug       bool
    http        HTTP
}

func (conn *CONN) GetContentLength() int {
    ret := 0

    if conn.Protocol == "http" {
        if conn.Debug {
            fmt.Println("DEBUG: use http protocol")
        }
        conn.http.Debug = conn.Debug
        conn.http.Connect(conn.Host, conn.Port)
        conn.http.Get(conn.Path, 1, 0)
        conn.http.Response()
        ret = conn.http.GetContentLength()
    }

    return ret
}
