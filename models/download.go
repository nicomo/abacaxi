package models

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/nicomo/abacaxi/logger"
)

// CreateUnimarcFile creates the file to be exported
func CreateUnimarcFile(ebk Ebook, fname string) (int64, error) {

	// create dirs if they don't exist
	path := filepath.Join("static", "downloads")
	pathErr := os.MkdirAll(path, os.ModePerm)
	if pathErr != nil {
		logger.Error.Println(path)
	}

	// create file
	f, err := os.Create(filepath.Join(path, fname))
	if err != nil {
		logger.Error.Println(err)
		return 0, err
	}
	defer f.Close()

	// get a buffered writer and write to file
	w := bufio.NewWriter(f)
	_, writeHeaderErr := w.WriteString("<?xml version=\"1.0\"?>\n")
	if writeHeaderErr != nil {
		logger.Error.Println(writeHeaderErr)
		return 0, writeHeaderErr
	}
	_, writeRecordErr := w.WriteString(ebk.RecordUnimarc)
	if writeRecordErr != nil {
		logger.Error.Println(writeRecordErr)
		return 0, writeRecordErr
	}

	w.Flush() // flush the buffer

	// get & return the size of the written file
	fi, fileInfoErr := f.Stat()
	if fileInfoErr != nil {
		logger.Error.Println("couldn't get file info: ", fileInfoErr)
		return 0, fileInfoErr
	}
	fs := fi.Size()

	return fs, nil
}
