package logger

import (
	"io"
	"log"
	"os"
)

var (
	Debug *log.Logger // for debugging purposes only
	Info  *log.Logger // "anything goes" info logged to console
	Error *log.Logger // erros logged to stdout + file
)

func init() {

	// set up log parameters
	logfile, logfile_err := os.OpenFile("ERMetadataHub_errors.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if logfile_err != nil {
		log.Fatalln("Failed to open error log file: ", logfile_err)
	}
	Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Error = log.New(io.MultiWriter(logfile, os.Stderr), "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}