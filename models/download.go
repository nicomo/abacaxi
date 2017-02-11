package models

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/nicomo/abacaxi/logger"
)

// CreateUnimarcFile creates the file to be exported
func CreateUnimarcFile(ebks []Ebook, fname string) (int64, error) {

	// create dirs if they don't exist
	path := filepath.Join("static", "downloads")
	ErrPath := os.MkdirAll(path, os.ModePerm)
	if ErrPath != nil {
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
	_, ErrWriteHeader := w.WriteString("<?xml version=\"1.0\"?>\n")
	if ErrWriteHeader != nil {
		logger.Error.Println(ErrWriteHeader)
		return 0, ErrWriteHeader
	}

	// write each marc record in turn
	for _, ebk := range ebks {
		_, ErrWriteRecord := w.WriteString(ebk.RecordUnimarc)
		if ErrWriteRecord != nil {
			logger.Error.Println(ErrWriteRecord)
			return 0, ErrWriteRecord
		}
	}

	w.Flush() // flush the buffer

	// get & return the size of the written file
	fi, ErrFileInfo := f.Stat()
	if ErrFileInfo != nil {
		logger.Error.Println("couldn't get file info: ", ErrFileInfo)
		return 0, ErrFileInfo
	}
	fs := fi.Size()

	return fs, nil
}
