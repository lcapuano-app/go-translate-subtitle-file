package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	none  = 0
	err   = 10
	info  = 20
	debug = 90
	dflt  = 90
)

var level int

func SetLogger(logPath, logLevel string) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	logFile := filepath.Join(logPath, "srt_translator.log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		SetLogger("", logLevel)
		return
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	level = getLogLevel(logLevel)
	Info("SRT Translator started")
}

func Debug(v ...any) {
	if level <= debug {
		log.Println("[DEBUG]:", v)
	}
}

func Info(v ...any) {
	if level >= info {
		log.Println("[INFO]:", v)
	}
}

func Err(v ...any) {
	if level >= err {
		log.Println("[ERROR]:", v)
	}
}

func Panic(v ...any) {
	log.Panic("[Panic]:", v)
}

func getLogLevel(level string) int {
	switch level {
	case "NONE":
		return none
	case "DEBUG":
		return debug
	case "INFO":
		return info
	case "ERROR":
		return err
	case "ERR":
		return err
	default:
		return dflt
	}
}
