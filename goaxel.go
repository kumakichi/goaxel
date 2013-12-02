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
    "sync"
    "sort"
    "github.com/xiangzhai/goaxel/conn"
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
    chunkFileIndex  []int
    wg              sync.WaitGroup
    bar             *pb.ProgressBar
)

func init() {
    flag.UintVar(&connNum, "n", 3, "Specify maximum speed (bytes per second)")
    flag.StringVar(&outputFileName, "o", defaultOutputFileName, "Specify local output file")
    flag.StringVar(&userAgent, "U", appName, "Set user agent")
    flag.BoolVar(&debug, "d", false, "Debug")
}

func connCallback(n int) {
    bar.Add(n)
}

func startRoutine(range_from, range_to int, old_range_from int, chunksize int,
                  url string) {
    defer wg.Done()
    protocol, host, port, strPath, userName, passwd := parseUrl(url)
    conn := &conn.CONN{Protocol: protocol, Host: host, Port: port,
                       UserAgent: userAgent, UserName: userName,
                       Passwd: passwd, Path: strPath, Debug: debug,
                       Callback: connCallback}
    conn.Get(range_from, range_to, outputFileName, old_range_from, chunksize)
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
        prefix := outputFileName + ".part."
        pos := strings.Index(path, prefix)
        if pos != -1 {
            key, _ := strconv.Atoi(path[pos + len(prefix):])
            chunkFileIndex = append(chunkFileIndex, key)
        }
        return nil
    })
    if err != nil {
        fmt.Printf("ERROR:", err.Error())
        return
    }
    sort.Ints(chunkFileIndex)
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
    if len(chunkFileIndex) != 0 {
        fmt.Println("Found chunk files, continue downloading...")
        hasChunk    = true
        connNum     = uint(len(chunkFileIndex))
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
            chunkFileSize = int(fileSize(fmt.Sprintf("%s.part.%d",
                outputFileName, chunkFileIndex[i])))
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
    if len(chunkFileIndex) == 0 {
        travelChunk(".")
    }
    for _, v := range chunkFileIndex {
        chunkFileName := fmt.Sprintf("%s.part.%d", outputFileName, v)
        chunkFile, _  := os.Open(chunkFileName)
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
        os.Remove(chunkFileName)
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

    if acceptRange && connNum != 1 {
        wg.Add(int(connNum))
        splitWork()
    } else {
        wg.Add(1)
        fmt.Println("It does not accept range, use signal connection instead")
        go startRoutine(0, 0, 0, 0, urls[0])
    }
    bar.Start()
    wg.Wait()
    writeChunk(".")
}
