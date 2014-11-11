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

func (this *CONN) httpConnect() bool {
    this.http.Debug = this.Debug
    this.http.Protocol = this.Protocol
    this.http.UserAgent = this.UserAgent
    return this.http.Connect(this.Host, this.Port)
}

func (this *CONN) httpsConnect() bool {
    this.https.Debug = this.Debug
    this.https.Protocol = this.Protocol
    this.https.UserAgent = this.UserAgent
    return this.https.Connect(this.Host, this.Port)
}

func (this *CONN) ftpConnect() bool {
    this.ftp.Debug = this.Debug
    if this.ftp.Connect(this.Host, this.Port) == false { return false }
    if this.UserName == "" { this.UserName = "anonymous" }
    this.ftp.Login(this.UserName, this.Passwd)
    if this.ftp.Code == 530 {
        fmt.Println("ERROR: login failure")
        return false
    }
    this.ftp.Request("TYPE I")
    dir := path.Dir(this.Path)
    if dir != "/" { dir += "/" }
    this.ftp.Cwd(dir)
    return true
}

func (this *CONN) GetContentLength(fileName string) (length int, accept bool) {
    length = 0
    accept = false

    if this.Protocol == "http" {
        if this.httpConnect() == false { return }
        this.http.Get(this.Path, 1, 0)
        this.http.Response()
        length = this.http.GetContentLength()
        accept = this.http.IsAcceptRange()
    } else if this.Protocol == "https" {
        if this.httpsConnect() == false { return }
        this.https.Get(this.Path, 1, 0)
        this.https.Response()
        length = this.https.GetContentLength()
        accept = this.https.IsAcceptRange()
    } else if this.Protocol == "ftp" {
        if this.ftpConnect() == false { return }
        length = this.ftp.Size(fileName)
        accept = true
    }

    return
}

func (this *CONN) Get(range_from, range_to int, fileName string, old_range_from int, chunksize int) {
    if this.Protocol == "http" {
        if this.httpConnect() == false { return }
        this.http.Callback = this.Callback
        this.http.Get(this.Path, range_from, range_to)
        this.http.WriteToFile(fileName, old_range_from, chunksize)
    } else if this.Protocol == "https" {
        if this.httpsConnect() == false { return }
        this.https.Callback = this.Callback
        this.https.Get(this.Path, range_from, range_to)
        this.https.WriteToFile(fileName, old_range_from, chunksize)
    } else if this.Protocol == "ftp" {
        if this.ftpConnect() == false { return }
        this.ftp.Callback = this.Callback
        this.ftp.Pasv()
        newConn := this.ftp.NewConnect()
        this.ftp.Request(fmt.Sprintf("REST %d", range_from))
        this.ftp.Request("RETR " + fileName)
        this.ftp.WriteToFile(newConn, fileName, range_from, range_to, chunksize)
    }
}
