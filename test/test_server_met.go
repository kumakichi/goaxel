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
    "os"
    "github.com/kumakichi/goaxel/emule"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: test_server_met filePath")
        return
    }
    o := new(emule.ServerMet)
    //o.Debug = true
    o.OpenFile(os.Args[1])
    fmt.Println(len(o.Servers))
    for i := 0; i < len(o.Servers); i++ {
        fmt.Println(o.Servers[i])
    }
}
