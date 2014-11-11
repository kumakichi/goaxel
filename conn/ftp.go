// FTP Client for Google Go language.
// Author: smallfish <smallfish.xy@gmail.com>

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
	"net"
	"strconv"
	"strings"
    "os"
    "io"
)

type FTP struct {
	host        string
	port        int
	user        string
	passwd      string
	pasv        int
	cmd         string
	Code        int
	Message     string
	Debug       bool
	stream      []byte
	conn        net.Conn
	Error       error
    offset      int
    Callback    func(int)
}

func (this *FTP) debugInfo(s string) {
	if this.Debug {
		fmt.Println(s)
	}
}

func (this *FTP) NewConnect() net.Conn {
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", this.host, this.pasv))
    if err != nil {
        this.Error = err
        fmt.Println("ERROR:", this.Error.Error())
    }
    if this.Debug {
        this.cmd = fmt.Sprintf("NewConnect:%d", this.pasv)
    }
    return conn
}

func (this *FTP) Connect(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	this.conn, this.Error = net.Dial("tcp", addr)
	if this.Error != nil {
        fmt.Println("ERROR:", this.Error.Error())
        return false
    }
    this.Response()
	this.host = host
	this.port = port
    return true
}

func (this *FTP) Login(user, passwd string) {
	this.Request("USER " + user)
	this.Request("PASS " + passwd)
	this.user = user
	this.passwd = passwd
}

func (this *FTP) WriteToFile(conn net.Conn, fileName string, old_range_from, range_to, offset int) {
    needLen := range_to - old_range_from + 1
    this.offset = offset
    defer conn.Close()
    data := make([]byte, 102400)
    chunkName := fmt.Sprintf("%s.part.%d", fileName, old_range_from)
    f, err := os.OpenFile(chunkName, os.O_CREATE | os.O_WRONLY, 0664)
    defer f.Close()
    if err != nil { panic(err) }
    for {
        n, err := conn.Read(data)
        if err != nil {
            if err != io.EOF { panic(err) }
        }
        if this.offset + n > needLen {
            n = needLen - this.offset
        }
        f.WriteAt(data[:n], int64(this.offset))
        if this.Callback != nil {
            this.Callback(n)
        }
        this.offset += n
        if err == io.EOF || this.offset == needLen { return }
    }
    return
}

func (this *FTP) Response() (code int, message string) {
	ret := make([]byte, 1024)
	n, _ := this.conn.Read(ret)
	msg := string(ret[:n])
	code, _ = strconv.Atoi(msg[:3])
	message = msg[4 : len(msg)-2]
	this.debugInfo("<*cmd*> " + this.cmd)
	this.debugInfo(fmt.Sprintf("<*code*> %d", code))
	this.debugInfo("<*message*> " + message)
	return
}

func (this *FTP) Request(cmd string) {
	if this.conn == nil { return }
    this.conn.Write([]byte(cmd + "\r\n"))
	this.cmd = cmd
	this.Code, this.Message = this.Response()
	if cmd == "PASV" {
		start, end := strings.Index(this.Message, "("), strings.Index(this.Message, ")")
		s := strings.Split(this.Message[start:end], ",")
		l1, _ := strconv.Atoi(s[len(s)-2])
		l2, _ := strconv.Atoi(s[len(s)-1])
		this.pasv = l1*256 + l2
	}
}

func (this *FTP) Pasv() {
	this.Request("PASV")
}

func (this *FTP) Pwd() {
	this.Request("PWD")
}

func (this *FTP) Cwd(path string) {
	this.Request("CWD " + path)
}

func (this *FTP) Mkd(path string) {
	this.Request("MKD " + path)
}

func (this *FTP) Size(path string) (size int) {
	this.Request("SIZE " + path)
	size, _ = strconv.Atoi(this.Message)
	return
}

func (this *FTP) List() {
	this.Pasv()
	this.Request("LIST")
}

func (this *FTP) Stor(file string, data []byte) {
	this.Pasv()
	if data != nil {
		this.stream = data
	}
	this.Request("STOR " + file)
}

func (this *FTP) Quit() {
	this.Request("QUIT")
	this.conn.Close()
}
