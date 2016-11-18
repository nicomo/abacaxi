package controllers

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type CSVConf struct {
	nfields    int
	ititle     int
	iauthors   []int
	ipublisher int
	isbn       int
	eisbn      int
	ipubdate   int
	iurl       int
	ilang      int
}

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

func csvIO(filename string, packname string) {
	csvData, err := csvClean(filename)
	if err != nil {
		logger.Error.Println(err)
	}

	csvSaveProcessed(csvData)
}

/*
// csvConfig reads the config for parsing the csv file provided by a given vendor
func csvConf(vendor string) CSVConf {
	configFile, err := os.Open("../../conf_csv.json")
	if err != nil {
		logger.Error.Println("cannot open conf_csv.json")
	}
	decoder := json.NewDecoder(configFile)
	conf := CSVConf{}
	decoderErr := decoder.Decode(&conf)
	if err != nil {
		logger.Error.Println(decoderErr)
	}

	return conf
}
*/
// csvClean takes a csv file, checks for length, some mandated fields, etc. and cleans it up
func csvClean(filename string) ([]CSVRecord, error) {

	// open csv file
	csvFile, err := os.Open(filename)
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

			if value == "" { // if value is empty, move on to next field
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
					// clean isbn, remove spaces & dashes
					csvRecord.isbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
					isbnCount++
				case 6:
					csvRecord.eisbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
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
	logger.Info.Printf("successfully parsed %d lines from %s - CSV contained %d isbn and %d eisbn", len(csvData), filename, isbnCount, eisbnCount)
	logger.Info.Println("rejected lines ", rejectedLines)

	return csvData, nil
}

// csvSaveProcessed saves cleaned values to a new, clean csv file

func csvSaveProcessed(csvData []CSVRecord) {

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
	outputFilename := "./data/cyberlibris_processed_" + t.Format("20060102150405") + ".csv"
	fileOutput, err := os.Create(outputFilename)
	if err != nil {
		logger.Error.Println(err)
	}
	defer fileOutput.Close()

	// create output CSV writer
	w := csv.NewWriter(fileOutput)
	w.Comma = ';'

	// save to file
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}
