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
	Protocol  string
	Host      string
	Port      int
	UserAgent string
	UserName  string
	Passwd    string
	Path      string
	Debug     bool
	Callback  func(int)
	http      HTTP
	https     HTTPS
	ftp       FTP
}

type DownLoader interface {
	Get(url string, from, size int)
	Response() (code int, message string)
	IsAcceptRange() bool
	GetContentLength(path string) int
	SetConnOpt(debug bool, userAgent string)
	Connect(host string, port int) bool
}

func (c *CONN) httpConnect() bool {
	c.http.Debug = c.Debug
	c.http.UserAgent = c.UserAgent

	return c.http.Connect(c.Host, c.Port)
}

func (c *CONN) httpsConnect() bool {
	c.https.Debug = c.Debug
	c.https.UserAgent = c.UserAgent

	return c.https.Connect(c.Host, c.Port)
}

func (c *CONN) ftpConnect() bool {
	c.ftp.Debug = c.Debug
	if c.ftp.Connect(c.Host, c.Port) == false {
		return false
	}
	if c.UserName == "" {
		c.UserName = "anonymous"
	}
	c.ftp.Login(c.UserName, c.Passwd)
	if c.ftp.Code == 530 {
		fmt.Println("ERROR: login failure")
		return false
	}
	c.ftp.Request("TYPE I")
	dir := path.Dir(c.Path)
	if dir != "/" {
		dir += "/"
	}
	c.ftp.Cwd(dir)
	return true
}

func (c *CONN) GetContentLength(fileName string) (length int, accept bool) {
	length = 0
	accept = false
	var downLoader DownLoader

	switch c.Protocol {
	case "http":
		downLoader = &HTTP{}
	case "https":
		downLoader = &HTTPS{}
	case "ftp":
		downLoader = &FTP{}
	}

	downLoader.SetConnOpt(c.Debug, c.UserAgent)
	if downLoader.Connect(c.Host, c.Port) == false {
		return
	}

	switch c.Protocol {
	case "http", "https":
		downLoader.Get(c.Path, 0, 0)
		downLoader.Response()
	}
	length = downLoader.GetContentLength(fileName)
	accept = downLoader.IsAcceptRange()

	return
}

func (c *CONN) Get(rangeFrom, pieceSize, alreadyHas int, fileName string) {
	if c.Protocol == "http" {
		if c.httpConnect() == false {
			return
		}
		c.http.Callback = c.Callback
		c.http.Get(c.Path, rangeFrom+alreadyHas, pieceSize)
		c.http.WriteToFile(fileName, rangeFrom, alreadyHas)
	} else if c.Protocol == "https" {
		if c.httpsConnect() == false {
			return
		}
		c.https.Callback = c.Callback
		c.https.Get(c.Path, rangeFrom+alreadyHas, pieceSize)
		c.https.WriteToFile(fileName, rangeFrom, alreadyHas)
	} else if c.Protocol == "ftp" {
		if c.ftpConnect() == false {
			return
		}
		c.ftp.Callback = c.Callback
		c.ftp.Pasv()
		newConn := c.ftp.NewConnect()
		c.ftp.Request(fmt.Sprintf("REST %d", rangeFrom+alreadyHas))
		c.ftp.Request("RETR " + fileName)
		c.ftp.WriteToFile(newConn, fileName, rangeFrom+alreadyHas, alreadyHas)
	}
}
