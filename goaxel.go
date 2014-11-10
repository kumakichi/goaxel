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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"./conn"
	"github.com/cheggaaa/pb"
)

const (
	appName               string = "GoAxel"
	defaultOutputFileName string = "default"
)

var (
	connNum        int
	userAgent      string
	versionPrint   bool
	debug          bool
	urls           []string
	outputPath     string
	outputFileName string
	outputFile     *os.File
	contentLength  int
	acceptRange    bool
	chunkFiles     []string
	ch             chan int
	bar            *pb.ProgressBar
)

type SortString []string

func (s SortString) Len() int {
	return len(s)
}

func (s SortString) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortString) Less(i, j int) bool {
	strI := strings.Split(s[i], ".part.")
	strJ := strings.Split(s[j], ".part.")
	numI, _ := strconv.Atoi(strI[1])
	numJ, _ := strconv.Atoi(strJ[1])
	return numI < numJ
}

func init() {
	flag.IntVar(&connNum, "n", 3, "Specify the number of connections")
	flag.StringVar(&outputFileName, "o", defaultOutputFileName, "Specify output file name. If more than 1 url specified,this option will be ignored")
	flag.StringVar(&userAgent, "U", appName, "Set user agent")
	flag.BoolVar(&debug, "d", false, "Print debug infomation")
	flag.StringVar(&outputPath, "p", ".", "Specify output file path")
	flag.BoolVar(&versionPrint, "V", false, "Print version and copyright")
}

func connCallback(n int) {
	bar.Add(n)
}

func startRoutine(range_from, range_to, alreadyHas int, url string) {
	protocol, host, port, strPath, userName, passwd := parseUrl(url)
	conn := &conn.CONN{Protocol: protocol, Host: host, Port: port,
		UserAgent: userAgent, UserName: userName,
		Passwd: passwd, Path: strPath, Debug: debug,
		Callback: connCallback}
	conn.Get(range_from, range_to, alreadyHas, outputFileName)
	ch <- 1
}

/* TODO: parse url to get host, port, path, basename */
func parseUrl(strUrl string) (protocol string, host string, port int,
	strPath string, userName string, passwd string) {
	protocol = ""
	host = ""
	port = 0
	strPath = ""

	u, err := url.Parse(strUrl)
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}
	protocol = u.Scheme
	host = u.Host
	port = 80
	if protocol == "https" {
		port = 443
	} else if protocol == "ftp" {
		port = 21
	}
	userinfo := u.User
	if userinfo != nil {
		userName = userinfo.Username()
		passwd, _ = userinfo.Password()
	}
	strPath = u.Path
	if strPath == "" {
		strPath = "/"
	}
	pos := strings.Index(host, ":")
	if pos != -1 {
		port, _ = strconv.Atoi(host[pos+1:])
		host = host[0:pos]
	}
	conn := &conn.CONN{Protocol: protocol, Host: host, Port: port,
		UserAgent: userAgent, UserName: userName,
		Passwd: passwd, Path: strPath, Debug: debug}
	if outputFileName == defaultOutputFileName && path.Base(strPath) != "/" {
		outputFileName = path.Base(strPath)
	}
	contentLength, acceptRange = conn.GetContentLength(outputFileName)
	return
}

func partialFileWalker(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	}
	if strings.HasPrefix(path, outputFileName+".part.") {
		chunkFiles = append(chunkFiles, path)
	}
	return nil
}

func travelChunk() {
	if err := filepath.Walk(".", partialFileWalker); err != nil {
		fmt.Printf("ERROR:", err.Error())
		return
	}
	sort.Sort(SortString(chunkFiles))
}

func fileSize(fileName string) int64 {
	if fi, err := os.Stat(fileName); err == nil {
		return fi.Size()
	}
	return 0
}

func splitWork(url string) {
	var filepath string
	var startPos, remainder int

	eachPieceSize := contentLength / connNum
	if eachPieceSize != 0 {
		remainder = contentLength - eachPieceSize*connNum
	}

	for i := 0; i < connNum; i++ {
		startPos = i * eachPieceSize
		filepath = fmt.Sprintf("%s.part.%d", outputFileName, startPos)
		chunkFileSize := int(fileSize(filepath))
		if chunkFileSize > 0 {
			bar.Add(chunkFileSize)
		}

		if i == connNum-1 { //the last piece,down addtional 'remainder',eg. split 9 to 4 + (4+'1')
			eachPieceSize += remainder
		}
		go startRoutine(startPos, startPos+eachPieceSize-1, chunkFileSize, url)
	}
}

func writeChunk() {
	if len(chunkFiles) == 0 {
		travelChunk()
	}

	for _, v := range chunkFiles {
		chunkFile, _ := os.Open(v)
		defer chunkFile.Close()
		chunkReader := bufio.NewReader(chunkFile)
		chunkWriter := bufio.NewWriter(outputFile)

		buf := make([]byte, 1024)
		for {
			n, err := chunkReader.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}
			if _, err := chunkWriter.Write(buf[:n]); err != nil {
				panic(err)
			}
		}
		if err := chunkWriter.Flush(); err != nil {
			panic(err)
		}
		os.Remove(v)
	}
}

func downSingleFile(url string) bool {
	var err error

	bar = pb.New(contentLength)
	bar.ShowSpeed = true
	bar.Units = pb.U_BYTES
	defer bar.Finish()

	if debug {
		fmt.Println("DEBUG: output filename", outputFileName)
		fmt.Println("DEBUG: content length", contentLength)
	}

	outputFile, err = os.Create(outputFileName)
	if err != nil {
		log.Println("error create:", outputFile, ",link:", url)
		return false
	}
	defer outputFile.Close()

	ch = make(chan int)
	if acceptRange && connNum != 1 {
		splitWork(url)
	} else {
		fmt.Println("It does not accept range, use signal connection instead")
		go startRoutine(0, 0, 0, url)
	}

	bar.Start()
	for i := 0; i < connNum; i++ {
		<-ch
	}
	writeChunk()

	return true
}

func main() {

	if len(os.Args) == 1 {
		fmt.Println("Usage: goaxel [options] url1 [url2] [url...]")
		fmt.Printf("	For more information,type %s -h\n", os.Args[0])
		return
	}

	flag.Parse()
	urls = flag.Args()

	if versionPrint {
		fmt.Println(fmt.Sprintf("%s Version 1.1", appName))
		fmt.Println("Copyright (C) 2013 Leslie Zhai")
		fmt.Println("Copyright (C) 2014 kumakichi")
	}

	if len(urls) == 0 {
		fmt.Println("You must specify at least one url to download")
		return
	}

	if outputPath != "." {
		err := os.Chdir(outputPath)
		if err != nil {
			log.Fatal("Change Dir Failed :", outputPath, err)
		}
	}

	for i := 0; i < len(urls); i++ {
		if len(urls) > 1 { // more than 1 url,can not set ouputfile name
			outputFileName = defaultOutputFileName
		}
		chunkFiles = make([]string, 0)

		parseUrl(urls[i])
		downSingleFile(urls[i])
	}
}
