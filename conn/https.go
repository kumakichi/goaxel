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

func (https *HTTPS) WriteToFile(outputFileName string, old_range_from int, chunkSize int) {
	https.offset = chunkSize
	defer https.conn.Close()
	resp := ""
	data := make([]byte, 1)
	for i := 0; ; {
		data := make([]byte, 1)
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
		resp += string(data[:n])
	}
	if https.Debug {
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
		n, err := https.conn.Read(data)
		if err != nil {
			if err != io.EOF {
				https.Error = err
				fmt.Println("ERROR:", https.Error.Error())
				return
			}
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
	return
}

func (https *HTTPS) Get(url string, rangeFrom, pieceSize int) {
	https.AddHeader(fmt.Sprintf("GET %s HTTP/1.0", url))
	https.AddHeader(fmt.Sprintf("Host: %s", https.host))
	if pieceSize == 0 {
		https.AddHeader(fmt.Sprintf("Range: bytes=0-"))
	} else {
		https.AddHeader(fmt.Sprintf("Range: bytes=%d-%d", rangeFrom, rangeFrom+pieceSize-1))
	}
	https.AddHeader(fmt.Sprintf("User-Agent: %s", https.UserAgent))
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

func (https *HTTPS) SetConnOpt(debug bool, userAgent string) {
	https.Debug = debug
	https.UserAgent = userAgent
}
