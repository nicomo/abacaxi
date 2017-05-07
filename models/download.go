package models

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/nicomo/abacaxi/logger"
)

func createFile(fname string) (*os.File, error) {
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
		return nil, err
	}
	return f, nil
}

// CreateKbartFile creates csv file with KBART fields from records
func CreateKbartFile(records []Record, fname string) (int64, error) {

	logger.Debug.Printf("in CreateKbartFile, fname : %s", fname)

	f, err := createFile(fname)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// create a new writer and change default separator
	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()

	// write header to csv file
	kbartHeader := []string{
		"publication_title",
		"print_identifier",
		"online_identifier",
		"date_first_issue_online",
		"num_first_vol_online",
		"num_first_issue_online",
		"date_last_issue_online",
		"num_last_vol_online",
		"num_last_issue_online",
		"title_url",
		"first_author",
		"title_id",
		"embargo_info",
		"coverage_depth",
		"coverage_notes",
		"publisher_name",
	}
	if err := w.Write(kbartHeader); err != nil {
		return 0, err
	}

	for _, record := range records {
		if err := w.Write(recordToKbart(record)); err != nil {
			logger.Error.Printf("couldn't write to csv file: %v", err)
			continue
		}
	}

	return getFileSize(f), nil
}

// CreateUnimarcFile creates the file to be exported
func CreateUnimarcFile(records []Record, fname string) (int64, error) {

	f, err := createFile(fname)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// get a buffered writer and write to file
	w := bufio.NewWriter(f)
	_, ErrWriteHeader := w.WriteString("<?xml version=\"1.0\"?>\n")
	if ErrWriteHeader != nil {
		return 0, ErrWriteHeader
	}

	// write each marc record in turn
	for _, record := range records {
		_, ErrWriteRecord := w.WriteString(record.RecordUnimarc)
		if ErrWriteRecord != nil {
			return 0, ErrWriteRecord
		}
	}

	w.Flush() // flush the buffer

	return getFileSize(f), nil
}

func getFileSize(f *os.File) int64 {
	// get & return the size of the written file
	fi, err := f.Stat()
	if err != nil {
		logger.Error.Printf("couldn't get file info: %v", err)
		return 0
	}
	return fi.Size()
}
