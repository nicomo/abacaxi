package logger

import (
	"io"
	"log"
	"os"
)

var (
	// Debug is for debugging purposes only
	Debug *log.Logger
	// Info : "anything goes" info logged to console
	Info *log.Logger
	// Error : errors logged to stdout + file
	Error *log.Logger
)

func init() {

	// set up log parameters
	logfile, logfileErr := os.OpenFile("ERMetadataHub_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if logfileErr != nil {
		log.Fatalln("Failed to open error log file: ", logfileErr)
	}
	Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Info = log.New(io.MultiWriter(logfile, os.Stderr), "INFO: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	Error = log.New(io.MultiWriter(logfile, os.Stderr), "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
