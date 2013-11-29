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
)

func main() {
    data := make([]byte, 1024)
    data[0] = 'H'
    data[1] = 'e'
    data[2] = 'l'
    data[3] = 'l'
    data[4] = 'o'
    fmt.Println(data)
    fmt.Println(string(data))
    var tmp []byte
    for i:= 0; i < len(data); i++ {
        if data[i] != 0 {
            tmp = append(tmp, data[i])
        }
    }
    fmt.Println(tmp)
    f1, err := os.OpenFile("test1.txt", os.O_CREATE | os.O_WRONLY, 0664)
    if err != nil {
        fmt.Println("ERROR:", err.Error())
        return
    }
    f1.Write(data)
    f1, _ = os.OpenFile("test1.txt", os.O_CREATE | os.O_WRONLY, 0664)
    f1.WriteAt([]byte("end"), 1024)
    f2, err := os.Create("test2.txt")
    if err != nil {
        fmt.Println("ERROR:", err.Error())
        return
    }
    f2.Write(tmp)
}
