package main

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func initLog() {
	info := ioutil.Discard
	warn := ioutil.Discard
	err := ioutil.Discard
	if debug {
		info = os.Stdout
		warn = os.Stdout
		err = os.Stderr
	}

	Info = log.New(info,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warn,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(err,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
