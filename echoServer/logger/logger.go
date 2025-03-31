package logger

import (
	"log"
	"os"
)

var infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

func Info(msg string) {
	infoLog.Println(msg)
}

func Warn(msg string) {
	warningLog.Println(msg)
}

func ErrorAndExit(msg string) {
	logAndExit(errorLog, msg)
}

func logAndExit(logger *log.Logger, msg string) {
	logger.Println(msg)
	os.Exit(1)
}
