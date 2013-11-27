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

type CONN struct {
    Protocol    string
    Host        string
    Port        int
    UserAgent   string
    User        string
    Passwd      string
    Path        string
    Debug       bool
    Callback    func(int)
    http        HTTP
}

func (conn *CONN) GetContentLength() (int, bool) {
    var length int = 0
    var accept bool = false

    if conn.Protocol == "http" {
        conn.http.Debug = conn.Debug
        conn.http.UserAgent = conn.UserAgent
        conn.http.Connect(conn.Host, conn.Port)
        conn.http.Get(conn.Path)
        conn.http.Response()
        length = conn.http.GetContentLength()
        accept = conn.http.IsAcceptRange()
    }

    return length, accept
}

func (conn *CONN) Get(range_from, range_to int, outputFileName string) {
    if conn.Protocol == "http" {
        conn.http.Debug = conn.Debug
        conn.http.UserAgent = conn.UserAgent
        conn.http.Callback = conn.Callback
        conn.http.Connect(conn.Host, conn.Port)
        conn.http.Get(conn.Path)
        conn.http.WriteToFile(range_from, range_to, outputFileName)
    }
}
