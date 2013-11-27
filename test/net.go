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
    "net"
    "regexp"
)

const (
    BUFFER_SIZE int = 1024
)

func main() {
    /* socket connect */
    conn, err := net.Dial("tcp", "www.linuxdeepin.com:80")
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /* socket write */
    _, err = conn.Write([]byte("GET /img/index_dau.png HTTP/1.0\r\nHost: www.linuxdeepin.com\r\nRange: bytes=1-\r\nUser-Agent: GoAxel 1.0\r\n\r\n"))
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        return;
    }

    /* socket read */
    data := make([]byte, BUFFER_SIZE)
    _, err = conn.Read(data)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
        conn.Close()
        return;
    }
    s := string(data[:BUFFER_SIZE])
    fmt.Println("DEBUG: ", s)
    conn.Close()

    /* parse http header */
    r, err := regexp.Compile(`Content-Length: (.*)`)
    if err != nil {
        fmt.Println("ERROR: ", err.Error())
    }
    result := r.FindStringSubmatch(s)
    if len(result) != 0 {
        fmt.Println("DEBUG: content length ", result[1])
    }
}
