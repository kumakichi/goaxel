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
	"time"

	"github.com/cheggaaa/pb"
	"github.com/kumakichi/goaxel/conn"
)

const (
	appName               string = "GoAxel"
	defaultOutputFileName string = "default"
)

type goAxelUrl struct {
	protocol string
	port     int
	userName string
	passwd   string
	path     string
	host     string
}

var (
	connNum        int
	userAgent      string
	showVersion    bool
	debug          bool
	urls           []string
	outputPath     string
	outputFileName string
	outputFile     *os.File
	contentLength  int
	acceptRange    bool
	noticeDone     chan int
	bar            *pb.ProgressBar
	cookieFile     string
	usrDefHeader   string
	usrDefUser     string
	usrDefPwd      string
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
	flag.StringVar(&outputFileName, "o", defaultOutputFileName,
		`Specify output file name 
		If more than 1 url specified, this option will be ignored`)
	flag.StringVar(&userAgent, "U", appName, "Set user agent")
	flag.BoolVar(&debug, "d", false, "Print debug infomation")
	flag.StringVar(&outputPath, "p", ".", "Specify output file path")
	flag.BoolVar(&showVersion, "V", false, "Print version and copyright")
	flag.StringVar(&cookieFile, "load-cookies", "", `Cookie file in the format 
		originally used by Netscape's cookies.txt`)
	flag.StringVar(&usrDefHeader, "header", "", `comma seperated header string`)
	flag.StringVar(&usrDefUser, "user", "", "Specify username")
	flag.StringVar(&usrDefPwd, "pass", "", "Specify password")
}

func connCallback(n int) {
	bar.Add(n)
}

func startRoutine(rangeFrom, pieceSize, alreadyHas int,
	u goAxelUrl, c []conn.Cookie, h []conn.Header) {
	conn := &conn.CONN{Protocol: u.protocol, Host: u.host, Port: u.port,
		UserAgent: userAgent, UserName: u.userName, Passwd: u.passwd,
		Path: u.path, Debug: debug, Callback: connCallback, Cookie: c, Header: h}
	conn.Get(rangeFrom, pieceSize, alreadyHas, outputFileName)
	noticeDone <- 1
}

//http://ts.test.com/file/a.aac?fid=77&tid=88
func getFixedUrlPath(rawurl string, u *url.URL) string {
	s := strings.Split(rawurl, u.Host)
	return s[1]
}

func parseUrl(urlStr string) (g goAxelUrl, e error) {
	ports := map[string]int{"http": 80, "https": 443, "ftp": 21}

	if ok := strings.Contains(urlStr, "//"); ok != true {
		urlStr = "http://" + urlStr //scheme not specified,treat it as http
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		e = err
		return
	}

	g.protocol = u.Scheme
	g.port = ports[g.protocol]

	if userinfo := u.User; userinfo != nil {
		g.userName = userinfo.Username()
		g.passwd, _ = userinfo.Password()
	}

	if usrDefUser != "" {
		g.userName = usrDefUser
	}

	if usrDefPwd != "" {
		g.passwd = usrDefPwd
	}

	if g.path = getFixedUrlPath(urlStr, u); g.path == "" { // links like : http://www.google.com
		g.path = "/"
	} else if outputFileName == defaultOutputFileName {
		outputFileName = path.Base(g.path)
	}

	g.host = u.Host
	pos := strings.Index(g.host, ":")
	if pos != -1 { // user defined port
		g.port, _ = strconv.Atoi(g.host[pos+1:])
		g.host = g.host[0:pos]
	}

	return
}

func getChunkFilesList(outputName string) (partFiles []string, e error) {
	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, outputName+".part.") {
			partFiles = append(partFiles, path)
		}
		return nil
	})

	if err != nil {
		e = err
		fmt.Printf("ERROR:", err.Error())
		return
	}
	sort.Sort(SortString(partFiles))

	return
}

func fileSize(fileName string) int64 {
	if fi, err := os.Stat(fileName); err == nil {
		return fi.Size()
	}
	return 0
}

func divideAndDownload(u goAxelUrl, cookie []conn.Cookie, header []conn.Header) {
	var filepath string
	var startPos, remainder int

	if acceptRange == false || connNum == 1 { //need not split work
		go startRoutine(0, 0, 0, u, cookie, header)
		return
	}

	eachPieceSize := contentLength / connNum
	remainder = contentLength - eachPieceSize*connNum

	for i := 0; i < connNum; i++ {
		startPos = i * eachPieceSize
		filepath = fmt.Sprintf("%s.part.%d", outputFileName, startPos)

		alreadyHas := int(fileSize(filepath))
		if alreadyHas > 0 {
			bar.Add(alreadyHas)
		}

		//the last piece,down addtional 'remainder',eg. split 9 to 4 + (4+'1')
		if i == connNum-1 {
			eachPieceSize += remainder
		}
		go startRoutine(startPos, eachPieceSize, alreadyHas, u, cookie, header)
	}
}

func mergeChunkFiles() {
	var n int
	var err error
	var chunkFiles []string

	chunkFiles, err = getChunkFilesList(outputFileName)
	if err != nil {
		log.Fatal("Merge chunk files failed :", err.Error())
	}

	for _, v := range chunkFiles {
		chunkFile, _ := os.Open(v)
		defer chunkFile.Close()

		chunkReader := bufio.NewReader(chunkFile)
		chunkWriter := bufio.NewWriter(outputFile)
		buf := make([]byte, 1024*1024)

		for {
			n, err = chunkReader.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}
			if _, err = chunkWriter.Write(buf[:n]); err != nil {
				panic(err)
			}
		}

		if err = chunkWriter.Flush(); err != nil {
			panic(err)
		}
		os.Remove(v)
	}
}

func getContentLengthAcceptRange(u goAxelUrl, c []conn.Cookie,
	h []conn.Header, outputName string) (int, bool) {
	conn := &conn.CONN{Protocol: u.protocol, Host: u.host, Port: u.port,
		UserAgent: userAgent, UserName: u.userName, Passwd: u.passwd,
		Path: u.path, Debug: debug, Cookie: c, Header: h}

	return conn.GetContentLength(outputName)
}

func createProgressBar(length int) (bar *pb.ProgressBar) {
	bar = pb.New(length)
	bar.ShowSpeed = true
	bar.Units = pb.U_BYTES
	return
}

func parseCookieLine(s []byte, host string) (c conn.Cookie, ok bool) {
	line := strings.Split(string(s), "\t")
	if len(line) != 7 {
		return
	}

	if len(line[0]) < 3 {
		return
	}

	if line[0][0] == '#' {
		return
	}

	domain := line[0]
	if domain[0] == '.' {
		domain = domain[1:]
	}

	//Curl source says, quoting Andre Garcia: "flag: A TRUE/FALSE
	// value indicating if all machines within a given domain can
	// access the variable.  This value is set automatically by the
	// browser, depending on the value set for the domain."
	//domainFlag := line[1]

	//path := line[2]
	//security := line[3]
	expires := line[4]
	name := line[5]
	value := line[6]

	if false == strings.Contains(host, domain) {
		return
	}

	expiresTimestamp, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		fmt.Println("expires timestamp convert error :", err)
		return
	}

	if expiresTimestamp < time.Now().Unix() {
		return
	}

	c.Key = name
	c.Val = value
	ok = true
	return
}

func loadUsrDefinedHeader(usrDef string) (header []conn.Header) {
	s := strings.Split(usrDef, ";")
	for i := 0; i < len(s); i++ {
		if len(s[i]) < 3 {
			continue
		}

		kv := strings.Split(s[i], ":")
		if len(kv) < 2 {
			continue
		}

		//Referer:http://www.google.com
		val := ""
		for j := 1; j < len(kv); j++ {
			val += kv[j]
			if j != len(kv)-1 {
				val += ":"
			}
		}
		header = append(header, conn.Header{Header: kv[0], Value: val})
	}
	return
}

func loadCookies(cookiePath, host string) (cookie []conn.Cookie, ok bool) {
	if cookiePath == "" {
		ok = false
		return
	}

	f, err := os.Open(cookiePath)
	if err != nil {
		fmt.Println("ERROR OPEN COOKIE :", err)
		ok = false
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		line, isPrefix, err := br.ReadLine()
		if err != nil && err != io.EOF {
			ok = false
			return
		}
		if isPrefix {
			fmt.Println("you should not see this message")
			continue
		}

		if c, ok := parseCookieLine(line, host); ok {
			cookie = append(cookie, c)
		}

		if err == io.EOF {
			break
		}
	}

	ok = true
	return
}

func downSingleFile(url string) bool {
	var err error

	u, err := parseUrl(url)
	if err != nil {
		return false
	}

	cookies, ok := loadCookies(cookieFile, u.host)
	if !ok {
		cookies = make([]conn.Cookie, 0)
	}

	headers := loadUsrDefinedHeader(usrDefHeader)

	contentLength, acceptRange = getContentLengthAcceptRange(u, cookies, headers, outputFileName)

	if debug {
		fmt.Printf("[DEBUG] content length:%d,accept range:%t, cookie file:%s\n",
			contentLength, acceptRange, cookieFile)
	}

	bar = createProgressBar(contentLength)
	defer bar.Finish()

	if outputFile, err = os.Create(outputFileName); err != nil {
		log.Println("error create:", outputFile, ",link:", url)
		return false
	}
	defer outputFile.Close()

	bar.Start()
	divideAndDownload(u, cookies, headers)

	for i := 0; i < connNum; i++ {
		<-noticeDone
	}
	mergeChunkFiles()

	return true
}

func showVersionInfo() {
	fmt.Println(fmt.Sprintf("%s Version 1.1", appName))
	fmt.Println("Copyright (C) 2013 Leslie Zhai")
	fmt.Println("Copyright (C) 2014 kumakichi")
}

func showUsage() {
	fmt.Println("Usage: goaxel [options] url1 [url2] [url...]")
	fmt.Printf("	For more information,type %s -h\n", os.Args[0])
}

func checkUrls(u *[]string) {
	if len(urls) == 0 {
		if false == showVersion {
			log.Fatal("You must specify at least one url to download")
		}
	}

	if len(urls) > 1 { // more than 1 url,can not set ouputfile name
		outputFileName = defaultOutputFileName
	}
}

func changeToOutputDir(dst string) {
	if dst != "." {
		if err := os.Chdir(dst); err != nil {
			log.Fatal("Change directory failed :", dst, err)
		}
	}
}

func getCookieAbsolutePath() {
	var err error

	if cookieFile == "" {
		return
	}

	cookieFile, err = filepath.Abs(cookieFile)
	if err != nil {
		cookieFile = ""
		log.Fatal("Error get absolute path :", cookieFile)
	}
}

func main() {
	if len(os.Args) == 1 {
		showUsage()
		return
	}

	flag.Parse()

	if showVersion {
		showVersionInfo()
	}

	urls = flag.Args()
	checkUrls(&urls)

	getCookieAbsolutePath()
	changeToOutputDir(outputPath)

	noticeDone = make(chan int)
	for i := 0; i < len(urls); i++ {
		downSingleFile(urls[i])
	}
}
