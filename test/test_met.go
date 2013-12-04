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
    "io/ioutil"
    "bytes"
    "encoding/binary"
)

func read_int16(data []byte) (ret int16) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func read_int32(data []byte) (ret int32) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func read_ip(data []byte) (ret string) {
    ret = fmt.Sprintf("%v.%v.%v.%v", data[0], data[1], data[2], data[3])

    return
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: test_met filepath")
        return
    }
    b, err := ioutil.ReadFile(os.Args[1])
    if err != nil {
        fmt.Println("ERROR:", err.Error())
    }
    fmt.Printf("0x%02x\n", b[0])
    fmt.Println("SERVERs COUNT:", b[1], b[2], b[3], b[4], read_int32(b[1:5]))
    fmt.Println("IP:", b[5], b[6], b[7], b[8], read_ip(b[5:9]))
    fmt.Println("PORT:", b[9], b[10], read_int16(b[9:11]))
    fmt.Println("TAGs COUNT:", b[11], b[12], b[13], b[14])
    fmt.Printf("TAG KIND 0x%02x\n", b[15])
    fmt.Println("TAG NAME:", b[16], b[17])
    fmt.Printf("TAG COUNT 0x%02x\n", b[18])
    fmt.Println("TAG LEN", b[19], b[20], read_int16(b[19:21]))
    fmt.Println(string(b[21:36]))
    fmt.Printf("TAG KIND 0x%02x\n", b[36])
    fmt.Println(b[37], b[38])
    fmt.Println(b[39])
    fmt.Println("TAG LEN", b[40], b[41], read_int16(b[40:42]))
    fmt.Println(string(b[42:58]))
    fmt.Printf("0x%02x\n", b[58])
    fmt.Println(b[59], b[60])
    fmt.Println(b[61])
    fmt.Println(b[62], b[63], b[64], b[65], read_int32(b[62:66]))
    fmt.Printf("0x%02x\n", b[66])
    fmt.Println(b[67], b[68])
    fmt.Println(b[69])
    fmt.Println(read_int32(b[70:74]))
}
