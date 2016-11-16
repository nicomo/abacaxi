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

	// open csv file
	csvFile, err := os.Open("../../data/cyberlibris_100.csv")
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

	// counter to keep track of records parsed, for logging
	line := 1

	for {
		// read a row
		record, err := reader.Read()
		// if at EOF, break out of loop
		if err == io.EOF {
			break
		}
		if len(record) != 10 {
			log.Printf("parsing line %d failed: invalid length of %d, expected 10\n", line, len(record))
		}

		// one record for the row
		var csvRecord CSVRecord

		for i, value := range record {
			authors := make([]string, 3) // replace 3 by length as recorded in config json
			switch i {
			case 0:
				csvRecord.publisher = value
			case 1:
				csvRecord.title = value
			case 2, 3, 4:
				authors = append(authors, value)
				csvRecord.authors = authors
			case 5:
				csvRecord.isbn = value
			case 6:
				csvRecord.eisbn = value
			case 7:
				csvRecord.pubdate = value
			case 8:
				csvRecord.url = value
			case 9:
				csvRecord.lang = value
			}

		}

		// add this particular record to the slice
		csvData = append(csvData, csvRecord)

		// increment line counter
		line++
	}

	// log number of records successfully parsed
	log.Printf("successfully parsed %d lines", len(csvData))

}
