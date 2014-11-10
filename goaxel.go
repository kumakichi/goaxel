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
    "fmt"
    "flag"
    "os"
    "net/url"
    "strings"
    "strconv"
    "path"
    "path/filepath"
    "io"
    "bufio"
    "sort"
    "github.com/kumakichi/goaxel/conn"
    "github.com/cheggaaa/pb"
)

const (
    appName                 string = "GoAxel"
    defaultOutputFileName   string = "default"
)

var (
    connNum         uint
    userAgent       string
    debug           bool
    urls            []string
    outputFileName  string
    outputFile      *os.File
    contentLength   int
    acceptRange     bool
    chunkFiles      []string
    ch              chan int
    bar             *pb.ProgressBar
)

type SortString []string

func (s SortString) Len() int {
    return len(s)
}

func (s SortString) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s SortString) Less(i,j int) bool {
    strI := strings.Split(s[i],".part.")
    strJ := strings.Split(s[j],".part.")
    numI,_ := strconv.Atoi(strI[1])
    numJ,_ := strconv.Atoi(strJ[1])
    return numI<numJ
}

func init() {
    flag.UintVar(&connNum, "n", 3, "Specify the number of connections")
    flag.StringVar(&outputFileName, "o", defaultOutputFileName, "Specify local output file")
    flag.StringVar(&userAgent, "U", appName, "Set user agent")
    flag.BoolVar(&debug, "d", false, "Debug")
}

func connCallback(n int) {
    bar.Add(n)
}

func startRoutine(range_from, range_to int, old_range_from int, chunksize int,
                  url string) {
    protocol, host, port, strPath, userName, passwd := parseUrl(url)
    conn := &conn.CONN{Protocol: protocol, Host: host, Port: port,
                       UserAgent: userAgent, UserName: userName,
                       Passwd: passwd, Path: strPath, Debug: debug,
                       Callback: connCallback}
    conn.Get(range_from, range_to, outputFileName, old_range_from, chunksize)
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
    if strPath == "" { strPath = "/" }
    pos := strings.Index(host, ":")
    if pos != -1 {
        port, _ = strconv.Atoi(host[pos + 1:])
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

func travelChunk(path string) {
    err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
        if f == nil { return err }
        if f.IsDir() { return nil }
        if strings.HasPrefix(path, outputFileName + ".part.") {
            chunkFiles = append(chunkFiles, path)
        }
        return nil
    })
    if err != nil {
        fmt.Printf("ERROR:", err.Error())
        return
    }
    sort.Sort(SortString(chunkFiles))
}

func fileSize(fileName string) (ret int64) {
    ret = 0
    f, err := os.Open(fileName)
    defer f.Close()
    if err != nil {
        panic(err)
        return
    }
    fi, _ := f.Stat()
    ret = fi.Size()
    return
}

func splitWork() {
    travelChunk(".")
    hasChunk := false
    if len(chunkFiles) != 0 {
        fmt.Println("Found chunk files, continue downloading...")
        hasChunk    = true
        connNum     = uint(len(chunkFiles))
    }
    offset := contentLength / int(connNum)
    remainder := 0
    if offset != 0 {
        remainder = contentLength % (offset * int(connNum))
    }
    start := 0
    for i := 0; i < int(connNum); i++ {
        chunkFileSize := 0
        if hasChunk {
            chunkFileSize = int(fileSize(chunkFiles[i]))
            bar.Add(chunkFileSize)
        }
        if i > len(urls) - 1 {
            go startRoutine(start + chunkFileSize, start + offset - 1, start,
                chunkFileSize, urls[len(urls) - 1])
        } else {
            go startRoutine(start + chunkFileSize, start + offset - 1, start,
                chunkFileSize, urls[i])
        }
        start += offset
        if (i == int(connNum) - 2) {
            offset += remainder
        }
    }
}

func writeChunk(path string) {
    if len(chunkFiles) == 0 {
        travelChunk(".")
    }
    for _, v := range chunkFiles {
        chunkFile, _  := os.Open(v)
        defer chunkFile.Close()
        chunkReader := bufio.NewReader(chunkFile)
        chunkWriter := bufio.NewWriter(outputFile)

        buf := make([]byte, 1024)
        for {
            n, err := chunkReader.Read(buf)
            if err != nil && err != io.EOF { panic(err) }
            if n == 0 { break }
            if _, err := chunkWriter.Write(buf[:n]); err != nil {
                panic(err)
            }
        }
        if err := chunkWriter.Flush(); err != nil { panic(err) }
        os.Remove(v)
    }
}

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: goaxel [options] url1 [url2] [url...]")
        return
    }

    if os.Args[1] == "-V" {
        fmt.Println(fmt.Sprintf("%s Version 1.0", appName))
        fmt.Println("Copyright 2013 Leslie Zhai")
        return
    }

    flag.Parse()

    for i := 0; i < len(flag.Args()); i++ {
        urls = append(urls, flag.Args()[i])
    }
    if len(urls) == 0 {
        fmt.Println("Invalid urls")
        return
    }
    /* TODO: mirror support */
    for i := 0; i < len(urls); i++ {
        parseUrl(urls[i])
    }

    bar = pb.New(contentLength)
    bar.ShowSpeed = true
    bar.Units = pb.U_BYTES
    if debug {
        fmt.Println("DEBUG: output filename", outputFileName)
        fmt.Println("DEBUG: content length", contentLength)
    }

    outputFile, _ = os.Create(outputFileName)
    defer outputFile.Close()

    ch = make(chan int)
    if acceptRange && connNum != 1 {
        splitWork()
    } else {
        fmt.Println("It does not accept range, use signal connection instead")
        go startRoutine(0, 0, 0, 0, urls[0])
    }
    bar.Start()
    for i := 0; i < int(connNum); i++ {
        <-ch
    }
    writeChunk(".")
}
