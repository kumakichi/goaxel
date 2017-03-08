package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func cfgContentProcess(link string, deleteMode bool) (info []string, contentBytes []byte) {
	md5str := link2md5(link)
	path := cfgPath + "/" + CFG_FILENAME

	b, err := ioutil.ReadFile(path)
	if nil != err {
		if !os.IsNotExist(err) {
			Warning.Println(err.Error())
		}
		return
	} else {
		content := string(b)
		lines := strings.Split(content, "\n")
		content = ""
		for _, v := range lines {
			arr := strings.Split(v, CFG_DELIMETER)
			if 4 != len(arr) {
				continue
			}

			if md5str == arr[0] {
				info = arr
				contentBytes = b
				if !deleteMode {
					break
				}
			}

			if deleteMode && md5str != arr[0] {
				content += v + "\n"
			}
		}
		if deleteMode {
			contentBytes = []byte(content)
		} else {
			contentBytes = b
		}
	}

	return
}

func cfgWrite(link string, length int, acceptRange bool, filename string) {
	md5str := link2md5(link)
	path := cfgPath + "/" + CFG_FILENAME

	arr, b := cfgContentProcess(link, false)
	if nil != arr {
		return
	}
	ar := 0
	if acceptRange {
		ar = 1
	}
	b = append(b, fmt.Sprintf("%s##%d##%d##%s\n", md5str, length, ar, filename)...)
	err := ioutil.WriteFile(path, b, 0600)
	if nil != err {
		Warning.Println(err.Error())
	}
}

func cfgRead(link string) (length int, acceptRange bool, filename string) {
	var err error
	length = -1
	filename = defaultOutputFileName

	arr, _ := cfgContentProcess(link, false)
	if nil == arr {
		return
	}

	length, err = strconv.Atoi(arr[1])
	if nil != err {
		Warning.Println(err.Error())
	}
	acceptRange = (arr[2] == "1")
	filename = arr[3]
	return
}

func cfgDelete(link string) {
	path := cfgPath + "/" + CFG_FILENAME
	_, b := cfgContentProcess(link, true)
	if len(b) > 0 {
		ioutil.WriteFile(path, b, 0600)
	} else {
		os.Remove(path)
	}
}

func link2md5(link string) string {
	Md5Inst := md5.New()
	Md5Inst.Write([]byte(link))
	Result := Md5Inst.Sum([]byte(""))
	return fmt.Sprintf("%x", Result)
}
