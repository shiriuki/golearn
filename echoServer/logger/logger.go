// Basic logger package. That allow to use log levels: Info, Warn, ErrorAndExit
package logger

import (
	"log"
	"os"
)

var infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

// Writes an Info log
func Info(msg string) {
	infoLog.Println(msg)
}

// Writes an Warn log
func Warn(msg string) {
	warningLog.Println(msg)
}

// Writes an Error log and terminates the application.
func ErrorAndExit(msg string) {
	logAndExit(errorLog, msg)
}

func logAndExit(logger *log.Logger, msg string) {
	logger.Println(msg)
	os.Exit(1)
}
