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
    "github.com/kumakichi/goaxel/conn"
)

const (
    outputFileName string = "test.png"
)

var (
    file            *os.File
    contentLength   int = 0
    received        int = 0
)

func callBack(n int) {
    received += n
    fmt.Println("received:", received)
}

func ftp_download(f *conn.FTP, path string) {
    conn := f.NewConnect()
    f.Request("REST 0")
    f.Request("RETR " + path)
    f.WriteToFile(conn, file, 0)
    return
}

func main() {
    file, _ = os.Create(outputFileName)
    defer file.Close()

    f := new(conn.FTP)
    defer f.Quit()
    f.Debug = true
    f.Callback = callBack
    f.Connect("localhost", 21)
    f.Login("anonymous", "")
    if f.Code == 530 {
        fmt.Println("ERROR: login failure")
        return
    }
    f.Request("TYPE I")
    f.Cwd("/")
    contentLength = f.Size(outputFileName)
    f.Pasv()
    ftp_download(f, outputFileName)
    return
}
