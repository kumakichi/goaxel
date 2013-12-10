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

package emule

import (
    "fmt"
    "log"
    "os"
    "bytes"
    "encoding/binary"
)

func byteToInt32(data []byte) (ret int32) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func byteToInt16(data []byte) (ret int16) {
    buf := bytes.NewBuffer(data)
    binary.Read(buf, binary.LittleEndian, &ret)
    return
}

func int16ToByte(data int16) (ret []byte) {
    ret = []byte{}
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, data)
    ret = buf.Bytes()
    return
}

func int32ToByte(data int32) (ret []byte) {
    ret = []byte{}
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, data)
    ret = buf.Bytes()
    return
}

func guid() (b []byte, uuid string) {
    b       = make([]byte, 16)
    uuid    = ""

    f, err := os.Open("/dev/urandom")
    if err != nil {
        log.Println(err.Error())
        return
    }
    defer f.Close()
    f.Read(b)
    uuid = fmt.Sprintf("%x-%x-%x-%x-%x-%x-%x-%x",
        b[0:2], b[2:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:14], b[14:16])
    return
}
