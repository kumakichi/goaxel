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
    "os"
)

func main() {
    println("Starting the server")

    listener, err := net.Listen("tcp", "0.0.0.0:7111")
    if err != nil {
        println("error listening:", err.Error())
        os.Exit(1)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            println("Error accept:", err.Error())
            return
        }
        go respConn(conn)
    }
}

func respConn(conn net.Conn) {
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        println("ERROR:", err.Error())
        return
    }
    fmt.Println("DEBUG:", buf[:n])
    return
}
