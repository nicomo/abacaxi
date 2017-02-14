package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// CSVRecord store one line from the csv file
type CSVRecord struct {
	authors   []string
	title     string
	isbn      string
	eisbn     string
	pubdate   string
	publisher string
	url       string
	lang      string
}

// csvIO takes a csv file to clean it, save copy & unmarshall content
func csvIO(filename string, tsname string, userM UserMessages) ([]models.Ebook, models.TargetService, UserMessages, error) {

	// retrieve target service (i.e. ebook package) for this file
	myTargetService, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// update date for TS publisher last harvest since
	// we're harvesting books from a publisher provided csv file
	myTargetService.TSPublisherLastHarvest = time.Now()

	// clean the csv file
	csvData, userM, err := csvClean(filename, myTargetService.TSCsvConf, userM)
	if err != nil {
		logger.Error.Println(err)
		userM["err"] = err
		return nil, myTargetService, userM, err
	}

	// save cleaned copy of csv file
	userM, ErrCsvSaveProcessed := csvSaveProcessed(csvData, tsname, userM)
	if ErrCsvSaveProcessed != nil {
		logger.Error.Println("couldn't save processed CSV", ErrCsvSaveProcessed)
		userM["err"] = ErrCsvSaveProcessed
		return nil, myTargetService, userM, ErrCsvSaveProcessed
	}

	// unmarshall csv records into ebook structs
	ebooks := []models.Ebook{}
	for _, record := range csvData {
		ebook := csvUnmarshall(record, myTargetService)
		ebooks = append(ebooks, ebook)
	}

	return ebooks, myTargetService, userM, nil
}

// csvClean takes a csv file, checks for length, some mandated fields, etc. and cleans it up
// FIXME: cyclomatic complexity 20 of function csvClean() is high (> 15) (gocyclo)
func csvClean(filename string, csvConf models.TSCSVConf, userM UserMessages) ([]CSVRecord, UserMessages, error) {

	// open csv file
	csvFile, err := os.Open(filename)
	if err != nil {
		logger.Error.Println("cannot open csv file")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// slice will hold successfully parsed records
	var csvData []CSVRecord

	// package csv has n fields, separator is ;
	reader.FieldsPerRecord = csvConfGetNFields(csvConf)
	reader.Comma = ';'

	// counters to keep track of records parsed, for logging
	line := 1
	var rejectedLines []int
	isbnCount := 1
	eisbnCount := 1

	for {
		// read a row
		record, err := reader.Read()

		// if at EOF, break out of loop
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error.Println(err)
			panic(err)
		}

		// if row not of the expected length, move on
		if len(record) != reader.FieldsPerRecord {
			logger.Info.Printf("parsing line %d failed: invalid length of %d, expected 10\n", line, len(record))
			rejectedLines = append(rejectedLines, line)
		}

		// one record for the row
		var csvRecord CSVRecord

		// authors are in a slice of string
		authors := make([]string, 0)

		for i, value := range record {
			if value == "" { // if value is empty, move on to next field
				continue
			}

			if !utf8.ValidString(value) { // encoding issue, not utf-8
				logger.Info.Printf("parsing line %d failed: not utf-8 encoded", line)
				break
			}

			// value not empty, save in struct
			switch i {
			// csvConf is indexed from 1, 0 being the nil value
			case csvConf.Publisher - 1:
				csvRecord.publisher = value
			case csvConf.Title - 1:
				csvRecord.title = value
			case csvConf.Isbn - 1:
				// clean isbn, remove spaces & dashes
				csvRecord.isbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
				isbnCount++
			case csvConf.Eisbn - 1:
				csvRecord.eisbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
				eisbnCount++
			case csvConf.Pubdate - 1:
				csvRecord.pubdate = value
			case csvConf.URL - 1:
				csvRecord.url = value
			case csvConf.Lang - 1:
				csvRecord.lang = value
			}

			// authors are in a slice
			for j := 0; j < len(csvConf.Authors); j++ {
				if i+1 == csvConf.Authors[j] {
					authors = append(authors, value)
				}
			}
		}

		// write the authors slice to the record
		csvRecord.authors = authors

		// if the record has neither an isbn nor an eisbn, not worth saving
		if csvRecord.eisbn == "" && csvRecord.isbn == "" {
			rejectedLines = append(rejectedLines, line)
			continue
		}

		// add this particular record to the slice
		csvData = append(csvData, csvRecord)

		// increment line counter
		line++
	}

	// log number of records successfully parsed
	parsedLog := fmt.Sprintf("successfully parsed %d lines from %s - CSV contained %d isbn and %d eisbn", len(csvData), filename, isbnCount, eisbnCount)
	logger.Info.Print(parsedLog)
	userM["parsedLog"] = parsedLog

	// log lines rejected
	if len(rejectedLines) > 0 {
		rejectedLinesLog := fmt.Sprintf("rejected lines in file %s: %v", filename, rejectedLines)
		logger.Info.Println(rejectedLinesLog)
		userM["rejectedLinesLog"] = rejectedLinesLog
	}

	return csvData, userM, nil
}

// csvSaveProcessed saves cleaned values to a new, clean csv file
// NOTE: this saves persistent data and should thus probably be in models.
func csvSaveProcessed(csvData []CSVRecord, tsname string, userM UserMessages) (UserMessages, error) {

	// change the []CSVRecord data into [][]string
	// so we can use encoding/csv to save to a cleaned up csv file
	var records [][]string
	for _, recordIn := range csvData {
		recordOut := make([]string, 0)
		for _, aut := range recordIn.authors {
			recordOut = append(recordOut, aut)
		}
		recordOut = append(recordOut,
			recordIn.title,
			recordIn.pubdate,
			recordIn.publisher,
			recordIn.isbn,
			recordIn.eisbn,
			recordIn.lang,
			recordIn.url)

		// append this record to the slice
		records = append(records, recordOut)
	}

	// create a new csv file
	t := time.Now()
	outputFilename := "./data/" + tsname + "Processed" + t.Format("20060102150405") + ".csv"
	fileOutput, err := os.Create(outputFilename)
	if err != nil {
		logger.Error.Println(err)
		return userM, err
	}
	defer fileOutput.Close()

	// create output CSV writer
	w := csv.NewWriter(fileOutput)
	w.Comma = ';'

	// save to file
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		logger.Error.Println(err)
		return userM, err
	}

	saveCopyMssg := fmt.Sprintf("successfully saved cleaned up version of csv file as %s", outputFilename)
	logger.Info.Println(saveCopyMssg)
	userM["saveCopyMssg"] = saveCopyMssg
	return userM, nil
}

// csvUnmarshall creates ebook object from csv record
func csvUnmarshall(recordIn CSVRecord, myTargetService models.TargetService) models.Ebook {
	ebk := models.Ebook{}
	for _, aut := range recordIn.authors {
		ebk.Authors = append(ebk.Authors, aut)
	}
	ebk.Publisher = recordIn.publisher
	if recordIn.isbn != "" {
		Isbn := models.Isbn{Isbn: recordIn.isbn, Electronic: false} // print isbn, not electronic
		ebk.Isbns = append(ebk.Isbns, Isbn)
	}
	if recordIn.eisbn != "" {
		Eisbn := models.Isbn{Isbn: recordIn.eisbn, Electronic: true} // eisbn, electronic
		ebk.Isbns = append(ebk.Isbns, Eisbn)
	}
	ebk.Title = recordIn.title
	ebk.Pubdate = recordIn.pubdate
	ebk.Lang = recordIn.lang
	ebk.PackageURL = recordIn.url
	ebk.PublisherLastHarvest = time.Now()
	ebk.TargetService = append(ebk.TargetService, myTargetService)
	if myTargetService.TSActive {
		ebk.Active = true
	}

	return ebk
}
