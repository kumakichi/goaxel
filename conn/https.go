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
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type HTTPS struct {
	host           string
	port           int
	user           string
	passwd         string
	Debug          bool
	UserAgent      string
	conn           *tls.Conn
	header         string
	headerResponse string
	offset         int
	Error          error
	Callback       func(int)
}

func (https *HTTPS) Connect(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	https.conn, https.Error = tls.Dial("tcp", address, nil)
	if https.Error != nil {
		fmt.Println("ERROR: ", https.Error.Error())
		return false
	}
	https.host = host
	https.port = port
	return true
}

func (https *HTTPS) AddHeader(header string) {
	https.header += header + "\r\n"
}

/* TODO: get this header */
func (https *HTTPS) Response() (code int, message string) {
	code = -1
	message = "NOT OK"
	defer https.conn.Close()
	data := make([]byte, 1)
	for i := 0; ; {
		n, err := https.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				https.Error = err
				fmt.Println("ERROR:", https.Error.Error())
				return
			}
		}
		if data[0] == '\r' {
			continue
		} else if data[0] == '\n' {
			if i == 0 {
				break
			}
			i = 0
		} else {
			i++
		}
		https.headerResponse += string(data[:n])
	}
	if https.Debug {
		fmt.Println("DEBUG:", https.headerResponse)
	}
	https.conn.Close()

	code = 200
	message = "OK"
	return
}

func (https *HTTPS) readResponseHeaders() {
	resp := make([]byte, 0)
	data := make([]byte, 1)
	comeAcrossLF := 0

	for {
		_, err := https.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				https.Error = err
				fmt.Println("ERROR:", https.Error.Error())
				return
			}
		}

		switch data[0] {
		case '\r':
			// do nothing
		case '\n':
			comeAcrossLF += 1
		default:
			comeAcrossLF = 0
		}

		if comeAcrossLF == 2 {
			break
		}
		resp = append(resp, data[0])
	}

	if https.Debug {
		fmt.Println("[DEBUG] RESP :", string(resp))
	}
}

func (https *HTTPS) writeContent(f *os.File) {
	data := make([]byte, buffer_size)

	for {
		n, err := https.conn.Read(data)
		if err != nil && err != io.EOF {
			https.Error = err
			fmt.Println("ERROR:", https.Error.Error())
			return
		}

		f.WriteAt(data[:n], int64(https.offset))
		if https.Callback != nil {
			https.Callback(n)
		}
		https.offset += n

		if err == io.EOF {
			return
		}
	}
}

func (https *HTTPS) WriteToFile(outputName string, rangeFrom,
	pieceSize, alreadyHas int) {
	https.offset = alreadyHas

	chunkName := fmt.Sprintf("%s.part.%d", outputName, rangeFrom)
	f, err := os.OpenFile(chunkName, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	https.readResponseHeaders()
	https.writeContent(f)

	return
}

func (https *HTTPS) loadUsrDefHeaders(h []Header) {
	if len(h) == 0 {
		return
	}

	for _, v := range h {
		https.AddHeader(fmt.Sprintf("%s:%s", v.Header, v.Value))
	}
}

func (https *HTTPS) loadCookies(c []Cookie) {
	if len(c) == 0 {
		return
	}

	cookie := ""
	for _, v := range c {
		cookie += v.Key + "=" + v.Val + ";"
	}
	cookie = cookie[:len(cookie)-1] //remove the last ';'
	https.AddHeader(fmt.Sprintf("Cookie:%s", cookie))
}

func (https *HTTPS) Get(url string, c []Cookie, h []Header, rangeFrom, pieceSize, alreadyHas int) {
	rangeFrom += alreadyHas

	https.AddHeader(fmt.Sprintf("GET %s HTTP/1.1", url))
	https.AddHeader("Connection: close") // default value is 'keep-alive'
	https.AddHeader(fmt.Sprintf("Host: %s", https.host))

	if pieceSize == 0 {
		https.AddHeader(fmt.Sprintf("Range: bytes=0-"))
	} else {
		https.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", rangeFrom, rangeFrom+pieceSize-1))
	}
	https.AddHeader(fmt.Sprintf("User-Agent: %s", https.UserAgent))
	https.loadUsrDefHeaders(h)
	https.loadCookies(c)
	https.AddHeader("")

	if https.Debug {
		fmt.Println("DEBUG:", https.header)
	}

	_, https.Error = https.conn.Write([]byte(https.header))
	if https.Error != nil {
		fmt.Println("ERROR: ", https.Error.Error())
	}
}

func (https *HTTPS) IsAcceptRange() bool {
	ret := false

	if strings.Contains(https.headerResponse, "Content-Range") ||
		strings.Contains(https.headerResponse, "Accept-Ranges") {
		ret = true
	}

	return ret
}

func (https *HTTPS) GetContentLength(junkVar string) int {
	ret := 0
	r, err := regexp.Compile(`Content-Length: (.*)`)
	if err != nil {
		https.Error = err
		fmt.Println("ERROR: ", err.Error())
		return ret
	}
	result := r.FindStringSubmatch(https.headerResponse)
	if len(result) != 0 {
		s := strings.TrimSuffix(result[1], "\r")
		ret, https.Error = strconv.Atoi(s)
	}
	return ret
}

func (https *HTTPS) SetConnOpt(c *CONN) {
	https.Debug = c.Debug
	https.UserAgent = c.UserAgent
}

func (https *HTTPS) SetCallBack(cb func(int)) {
	https.Callback = cb
}

func (https *HTTPS) Close() {
	https.conn.Close()
}
