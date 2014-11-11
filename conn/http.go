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
	Protocol       string
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

const (
	buffer_size int = 102400
)

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
	defer http.conn.Close()
	data := make([]byte, 1)
	for i := 0; ; {
		n, err := http.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				http.Error = err
				fmt.Println("ERROR:", http.Error.Error())
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
		http.headerResponse += string(data[:n])
	}
	if http.Debug {
		fmt.Println("DEBUG:", http.headerResponse)
	}
	http.conn.Close()

	code = 200
	message = "OK"
	return
}

func (http *HTTP) WriteToFile(outputFileName string, old_range_from int, chunkSize int) {
	http.offset = chunkSize
	defer http.conn.Close()
	resp := ""
	data := make([]byte, 1)
	for i := 0; ; {
		data := make([]byte, 1)
		n, err := http.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				http.Error = err
				fmt.Println("ERROR:", http.Error.Error())
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
		resp += string(data[:n])
	}
	if http.Debug {
		fmt.Println("DEBUG:", resp)
	}
	chunkName := fmt.Sprintf("%s.part.%d", outputFileName, old_range_from)
	f, err := os.OpenFile(chunkName, os.O_CREATE|os.O_WRONLY, 0664)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	data = make([]byte, buffer_size)
	for {
		n, err := http.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				http.Error = err
				fmt.Println("ERROR:", http.Error.Error())
				return
			}
		}
		f.WriteAt(data[:n], int64(http.offset))
		if http.Callback != nil {
			http.Callback(n)
		}
		http.offset += n
		if err == io.EOF {
			return
		}
	}
	return
}

func (http *HTTP) Get(url string, rangeFrom, pieceSize int) {
	http.AddHeader(fmt.Sprintf("GET %s HTTP/1.0", url))
	http.AddHeader(fmt.Sprintf("Host: %s", http.host))
	if pieceSize == 0 {
		http.AddHeader(fmt.Sprintf("Range: bytes=0-"))
	} else {
		http.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", rangeFrom, rangeFrom+pieceSize-1))
	}
	http.AddHeader(fmt.Sprintf("User-Agent: %s", http.UserAgent))
	http.AddHeader("")
	if http.Debug {
		fmt.Println("DEBUG:", http.header)
	}
	_, http.Error = http.conn.Write([]byte(http.header))
	if http.Error != nil {
		fmt.Println("ERROR: ", http.Error.Error())
	}
}

func (http *HTTP) IsAcceptRange() bool {
	ret := false

	if strings.Contains(http.headerResponse, "Content-Range") ||
		strings.Contains(http.headerResponse, "Accept-Ranges") {
		ret = true
	}

	return ret
}

func (http *HTTP) GetContentLength() int {
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
