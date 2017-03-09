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
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type HTTP struct {
	host           string
	port           int
	user           string
	passwd         string
	Debug          bool
	UserAgent      string
	conn           net.Conn
	header         string
	headerResponse string
	offset         int
	Error          error
	Callback       func(int)
}

func (http *HTTP) Connect(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	http.conn, http.Error = net.Dial("tcp", address)
	if http.Error != nil {
		fmt.Println("ERROR: ", http.Error.Error())
		return false
	}
	http.host = host
	http.port = port
	return true
}

func (http *HTTP) AddHeader(header string) {
	http.header += header + "\r\n"
}

/* TODO: get this header */
func (http *HTTP) Response() (code int, message string) {
	code = -1
	message = "NOT OK"
	//defer http.conn.Close()
	data := make([]byte, 1)
	for i := 0; ; {
		n, err := http.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				http.Error = err
				fmt.Println("ERROR:", http.Error.Error())
				return
			}
			break
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
		http.headerResponse += string(data[:n])
	}
	if http.Debug {
		fmt.Println("DEBUG:", http.headerResponse)
	}

	code = 200
	message = "OK"
	return
}

func (http *HTTP) readResponseHeaders() {
	resp := make([]byte, 0)
	data := make([]byte, 1)
	comeAcrossLF := 0

	for {
		_, err := http.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				http.Error = err
				fmt.Println("ERROR:", http.Error.Error())
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

	if http.Debug {
		fmt.Println("[DEBUG] RESP :", string(resp))
	}
}

func (http *HTTP) writeContent(f *os.File, pieceSize, alreadyHas int) (written int) {
	data := make([]byte, buffer_size)
	written = 0

	var n int
	var err error
	for {
		left := pieceSize - alreadyHas - written
		if left >= buffer_size {
			n, err = http.conn.Read(data)
		} else {
			n, err = http.conn.Read(data[:left])
		}

		if err != nil && err != io.EOF {
			http.Error = err
			fmt.Println("ERROR:", http.Error.Error())
			return
		}

		f.WriteAt(data[:n], int64(http.offset))
		if http.Callback != nil {
			http.Callback(n)
		}
		http.offset += n
		written += n

		if err == io.EOF || http.offset == pieceSize {
			return
		}
	}
}

func (http *HTTP) WriteToFile(outputName string, rangeFrom,
	pieceSize, alreadyHas int) int {
	http.offset = alreadyHas

	chunkName := fmt.Sprintf("%s.part.%d", outputName, rangeFrom)
	f, err := os.OpenFile(chunkName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	http.readResponseHeaders()
	return http.writeContent(f, pieceSize, alreadyHas)
}

func (http *HTTP) loadUsrDefHeaders(h []Header) {
	if len(h) == 0 {
		return
	}

	for _, v := range h {
		http.AddHeader(fmt.Sprintf("%s:%s", v.Header, v.Value))
	}
}

func (http *HTTP) loadCookies(c []Cookie) {
	if len(c) == 0 {
		return
	}

	cookie := ""
	for _, v := range c {
		cookie += v.Key + "=" + v.Val + ";"
	}
	cookie = cookie[:len(cookie)-1] //remove the last ';'
	http.AddHeader(fmt.Sprintf("Cookie:%s", cookie))
}

func (http *HTTP) Get(url string, c []Cookie, h []Header, rangeFrom, pieceSize, alreadyHas int) (err error) {
	rangeFrom += alreadyHas

	http.AddHeader(fmt.Sprintf("GET %s HTTP/1.1", url))
	http.AddHeader("Connection: close") // default value of HTTP1.1 is 'keep-alive'
	http.AddHeader(fmt.Sprintf("Host: %s", http.host))

	if pieceSize == 0 {
		http.AddHeader(fmt.Sprintf("Range: bytes=0-"))
	} else {
		http.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", rangeFrom, rangeFrom+pieceSize-1))
	}
	http.AddHeader(fmt.Sprintf("User-Agent: %s", http.UserAgent))
	http.loadUsrDefHeaders(h)
	http.loadCookies(c)
	http.AddHeader("")

	if http.Debug {
		fmt.Println("HEADER:", http.header)
	}

	_, http.Error = http.conn.Write([]byte(http.header))
	if http.Error != nil {
		err = http.Error
		fmt.Println("ERROR: ", http.Error.Error())
	}
	return
}

func (http *HTTP) GetFilename() string {
	r, _ := regexp.Compile(`filename="(.*)"`)
	result := r.FindStringSubmatch(http.headerResponse)
	if len(result) > 1 {
		return result[1]
	}
	return ""
}

func (http *HTTP) IsAcceptRange() bool {
	ret := false

	if strings.Contains(http.headerResponse, "Content-Range") ||
		strings.Contains(http.headerResponse, "Accept-Ranges") {
		ret = true
	}

	return ret
}

func (http *HTTP) GetContentLength(junkVar string) int {
	ret := 0
	r, err := regexp.Compile(`Content-Length: (.*)`)
	if err != nil {
		http.Error = err
		fmt.Println("ERROR: ", err.Error())
		return ret
	}
	result := r.FindStringSubmatch(http.headerResponse)
	if len(result) != 0 {
		s := strings.TrimSuffix(result[1], "\r")
		ret, http.Error = strconv.Atoi(s)
	}
	return ret
}

func (http *HTTP) SetConnOpt(c *CONN) {
	http.Debug = c.Debug
	http.UserAgent = c.UserAgent
}

func (http *HTTP) SetCallBack(cb func(int)) {
	http.Callback = cb
}

func (http *HTTP) Close() {
	http.conn.Close()
}
