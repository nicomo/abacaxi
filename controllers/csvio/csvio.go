package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

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

func main() {

	fileName := "../../data/cyberlibris_100.csv"
	// open csv file
	csvFile, err := os.Open(fileName)
	if err != nil {
		logger.Error.Println("cannot open csv file")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// slice will hold successfully parsed records
	var csvData []CSVRecord

	// cyberlibris has 10 fields, separator is ;
	reader.FieldsPerRecord = 10
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
		}

		if len(record) != 10 {
			log.Printf("parsing line %d failed: invalid length of %d, expected 10\n", line, len(record))
			rejectedLines = append(rejectedLines, line)
		}

		// one record for the row
		var csvRecord CSVRecord
		authors := make([]string, 0)

		for i, value := range record {

			if value == "" { // is value is empty, move on to next field
				continue
			} else { // value not empty, save in struct
				switch i {
				case 0:
					csvRecord.publisher = value
				case 1:
					csvRecord.title = value
				case 2, 3, 4:
					authors = append(authors, value)
				case 5:
					csvRecord.isbn = value
					isbnCount++
				case 6:
					csvRecord.eisbn = value
					eisbnCount++
				case 7:
					csvRecord.pubdate = value
				case 8:
					csvRecord.url = value
				case 9:
					csvRecord.lang = value
				}

			}

		}

		// write the authors slice to the record
		csvRecord.authors = authors

		// if the record doesn't have at least an isbn || eisbn, not worth saving
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
	logger.Info.Printf("successfully parsed %d lines from %s - CSV contained %d isbn and %d eisbn", len(csvData), fileName, isbnCount, eisbnCount)
	logger.Info.Println("rejected lines ", rejectedLines)
}
