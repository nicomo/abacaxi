package controllers

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
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

// csvIO takes a csv file to clean it, save copy & unmarshall content
func csvIO(filename string, packname string) ([]models.Ebook, error) {

	logger.Debug.Println(packname)

	// clean the csv file
	csvData, err := csvClean(filename)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}

	// save cleaned copy of csv file
	csvSaveProcessedErr := csvSaveProcessed(csvData)
	if csvSaveProcessedErr != nil {
		logger.Error.Println("couldn't save processed CSV", csvSaveProcessedErr)
		return nil, csvSaveProcessedErr
	}

	// unmarshall csv records into ebook structs
	ebooks := []models.Ebook{}
	for _, record := range csvData {
		ebook := csvUnmarshall(record, packname)
		ebooks = append(ebooks, ebook)
	}

	return ebooks, nil
}

// csvClean takes a csv file, checks for length, some mandated fields, etc. and cleans it up
func csvClean(filename string) ([]CSVRecord, error) {

	// open csv file
	csvFile, err := os.Open(filename)
	if err != nil {
		logger.Error.Println("cannot open csv file")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// is Valid utf-8?

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

		// if row not if the expected length, move on
		if len(record) != 10 {
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
			} else if !utf8.ValidString(value) { // encoding issue, not utf-8
				logger.Info.Printf("parsing line %d failed: not utf-8 encoded", line)
				break
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
func csvSaveProcessed(csvData []CSVRecord) error {

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
		return err
	}
	defer fileOutput.Close()

	// create output CSV writer
	w := csv.NewWriter(fileOutput)
	w.Comma = ';'

	// save to file
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		logger.Error.Println(err)
		return err
	}

	logger.Info.Printf("successfully saved cleaned up version of csv file as %s", outputFilename)
	return nil
}

// csvUnmarshall creates ebook object from csv record
func csvUnmarshall(recordIn CSVRecord, packname string) models.Ebook {
	ebk := models.Ebook{}
	for _, aut := range recordIn.authors {
		ebk.Authors = append(ebk.Authors, aut)
	}
	ebk.Publisher = recordIn.publisher
	Isbn := models.Isbn{recordIn.isbn, false, false} // print isbn, not electronic, not primary
	Eisbn := models.Isbn{recordIn.eisbn, true, true} // eisbn, electronic, primary
	ebk.Isbns = append(ebk.Isbns, Isbn, Eisbn)
	ebk.Title = recordIn.title
	ebk.Pubdate = recordIn.pubdate
	ebk.Lang = recordIn.lang
	ebk.PackageURL = recordIn.url
	ebk.PublisherLastHarvest = time.Now()
	ebk.TargetService = packname

	return ebk
}
