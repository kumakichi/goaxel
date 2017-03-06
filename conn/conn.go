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

type Cookie struct {
	Key, Val string
}

type Header struct {
	Header, Value string
}

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
	Cookie    []Cookie
	Header    []Header
}

type DownLoader interface {
	Get(url string, c []Cookie, h []Header, rangeFrom, pieceSize, alreadyHas int)
	Response() (code int, message string)
	IsAcceptRange() bool
	GetContentLength(path string) int
	SetConnOpt(c *CONN)
	Connect(host string, port int) bool
	SetCallBack(cb func(int))
	WriteToFile(fileName string, rangeFrom, pieceSize, alreadyHas int)
	Close()
}

const (
	buffer_size int = 1024 * 1024
)

func (c *CONN) Connect() (downLoader DownLoader, ok bool) {
	ok = false

	switch c.Protocol {
	case "http":
		downLoader = &HTTP{}
	case "https":
		downLoader = &HTTPS{}
	case "ftp":
		downLoader = &FTP{}
	}

	downLoader.SetConnOpt(c)
	if downLoader.Connect(c.Host, c.Port) == true {
		ok = true
	}

	return
}

func (c *CONN) GetContentLength(fileName string) (length int, accept bool, loader DownLoader) {
	length = 0
	accept = false

	downLoader, ok := c.Connect()
	if false == ok {
		return
	}
	//defer downLoader.Close()

	switch c.Protocol {
	case "http", "https":
		downLoader.Get(c.Path, c.Cookie, c.Header, 0, 0, 0)
		downLoader.Response()
	case "ftp":
	}
	length = downLoader.GetContentLength(fileName)
	accept = downLoader.IsAcceptRange()
	loader = downLoader

	return
}

func (c *CONN) Get(rangeFrom, pieceSize, alreadyHas int, fileName string) {
	downLoader, ok := c.Connect()
	if false == ok {
		return
	}
	defer downLoader.Close()

	downLoader.SetCallBack(c.Callback)
	downLoader.Get(c.Path, c.Cookie, c.Header, rangeFrom, pieceSize, alreadyHas)
	downLoader.WriteToFile(fileName, rangeFrom, pieceSize, alreadyHas)
}
