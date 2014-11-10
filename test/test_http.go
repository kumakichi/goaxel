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
    "github.com/kumakichi/goaxel/conn"
)

func main() {
    http := new(conn.HTTP)
    http.Debug = true
    http.Connect("localhost", 80)
    http.Get("/test.mp4", 1, 0)
    http.Response()
    length := http.GetContentLength()
    fmt.Println("content length: ", length)
    range_len := length / 6
    for i := 0; i < 6; i++ {
        if i != 5 {
            fmt.Printf("range %d - %d\n", 1 + i * range_len, range_len + i * range_len)
        } else {
            fmt.Printf("range %d - %d\n", 1 + i * range_len, length)
        }
    }
}
