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
    "path"
)

type CONN struct {
    Protocol    string
    Host        string
    Port        int
    UserAgent   string
    UserName    string
    Passwd      string
    Path        string
    Debug       bool
    Callback    func(int)
    http        HTTP
    https       HTTPS
    ftp         FTP
}

func (conn *CONN) httpConnect() bool {
    conn.http.Debug = conn.Debug
    conn.http.Protocol = conn.Protocol
    conn.http.UserAgent = conn.UserAgent
    return conn.http.Connect(conn.Host, conn.Port)
}

func (conn *CONN) httpsConnect() bool {
    conn.https.Debug = conn.Debug
    conn.https.Protocol = conn.Protocol
    conn.https.UserAgent = conn.UserAgent
    return conn.https.Connect(conn.Host, conn.Port)
}

func (conn *CONN) ftpConnect() bool {
    conn.ftp.Debug = conn.Debug
    if conn.ftp.Connect(conn.Host, conn.Port) == false { return false }
    if conn.UserName == "" { conn.UserName = "anonymous" }
    conn.ftp.Login(conn.UserName, conn.Passwd)
    if conn.ftp.Code == 530 {
        fmt.Println("ERROR: login failure")
        return false
    }
    conn.ftp.Request("TYPE I")
    dir := path.Dir(conn.Path)
    if dir != "/" { dir += "/" }
    conn.ftp.Cwd(dir)
    return true
}

func (conn *CONN) GetContentLength(fileName string) (length int, accept bool) {
    length = 0
    accept = false

    if conn.Protocol == "http" {
        if conn.httpConnect() == false { return }
        conn.http.Get(conn.Path, 1, 0)
        conn.http.Response()
        length = conn.http.GetContentLength()
        accept = conn.http.IsAcceptRange()
    } else if conn.Protocol == "https" {
        if conn.httpsConnect() == false { return }
        conn.https.Get(conn.Path, 1, 0)
        conn.https.Response()
        length = conn.https.GetContentLength()
        accept = conn.https.IsAcceptRange()
    } else if conn.Protocol == "ftp" {
        if conn.ftpConnect() == false { return }
        length = conn.ftp.Size(fileName)
        accept = true
    }

    return
}

func (conn *CONN) Get(range_from, range_to int, fileName string, old_range_from int, chunksize int) {
    if conn.Protocol == "http" {
        if conn.httpConnect() == false { return }
        conn.http.Callback = conn.Callback
        conn.http.Get(conn.Path, range_from, range_to)
        conn.http.WriteToFile(fileName, old_range_from, chunksize)
    } else if conn.Protocol == "https" {
        if conn.httpsConnect() == false { return }
        conn.https.Callback = conn.Callback
        conn.https.Get(conn.Path, range_from, range_to)
        conn.https.WriteToFile(fileName, old_range_from, chunksize)
    } else if conn.Protocol == "ftp" {
        if conn.ftpConnect() == false { return }
        conn.ftp.Callback = conn.Callback
        conn.ftp.Pasv()
        newConn := conn.ftp.NewConnect()
        conn.ftp.Request(fmt.Sprintf("REST %d", range_from))
        conn.ftp.Request("RETR " + fileName)
        conn.ftp.WriteToFile(newConn, fileName, range_from, chunksize)
    }
}
