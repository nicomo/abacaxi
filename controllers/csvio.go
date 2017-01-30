package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// CSVConf indicates the column (index) in the csv file
// where the various pieces of info are located
type CSVConf struct {
	Nfields    int   `json:"nfields"`
	Ititle     int   `json:"ititle"`
	Iauthors   []int `json:"iauthors"`
	Ipublisher int   `json:"ipublisher"`
	Isbn       int   `json:"isbn"`
	Eisbn      int   `json:"eisbn"`
	Ipubdate   int   `json:"ipubdate"`
	Iurl       int   `json:"iurl"`
	Ilang      int   `json:"ilang"`
}

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

// getCsvConf loads the conf for parsing a particular package csv
func getCsvConf(tsname string) (CSVConf, error) {

	// open & read the json csv conf file
	file, err := ioutil.ReadFile("./csv-conf/" + tsname + ".json")
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}

	// unmarshal json into a CSVConf
	csvConf := CSVConf{}
	jsonUnmarshalErr := json.Unmarshal(file, &csvConf)
	if jsonUnmarshalErr != nil {
		logger.Error.Println(jsonUnmarshalErr)
	}

	return csvConf, nil
}

// csvIO takes a csv file to clean it, save copy & unmarshall content
func csvIO(filename string, tsname string, userM userMessages) ([]models.Ebook, models.TargetService, userMessages, error) {

	// retrieve target service (i.e. ebook package) for this file
	myTargetService, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// update date for TS publisher last harvest since
	// we're harvesting books from a publisher provided csv file
	myTargetService.TSPublisherLastHarvest = time.Now()

	// load the config for this package from the csv_conf file
	csvConf, err := getCsvConf(tsname)
	if err != nil {
		logger.Error.Println(err)
		return nil, myTargetService, nil, err
	}

	// clean the csv file
	csvData, userM, err := csvClean(filename, csvConf, userM)
	if err != nil {
		logger.Error.Println(err)
		userM["err"] = err
		return nil, myTargetService, userM, err
	}

	// save cleaned copy of csv file
	userM, csvSaveProcessedErr := csvSaveProcessed(csvData, tsname, userM)
	if csvSaveProcessedErr != nil {
		logger.Error.Println("couldn't save processed CSV", csvSaveProcessedErr)
		userM["err"] = csvSaveProcessedErr
		return nil, myTargetService, userM, csvSaveProcessedErr
	}

	// unmarshall csv records into ebook structs
	//FIXME: creation of isbns is not correct, with results like [{ false} {9781607807582 true}]
	ebooks := []models.Ebook{}
	for _, record := range csvData {
		ebook := csvUnmarshall(record, myTargetService)
		ebooks = append(ebooks, ebook)
	}

	return ebooks, myTargetService, userM, nil
}

// csvClean takes a csv file, checks for length, some mandated fields, etc. and cleans it up
// FIXME: cyclomatic complexity 20 of function csvClean() is high (> 15) (gocyclo)
func csvClean(filename string, csvConf CSVConf, userM userMessages) ([]CSVRecord, userMessages, error) {

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
	reader.FieldsPerRecord = csvConf.Nfields
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

		// if row not if the expected length, move on
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
			} else if !utf8.ValidString(value) { // encoding issue, not utf-8
				logger.Info.Printf("parsing line %d failed: not utf-8 encoded", line)
				break
			} else { // value not empty, save in struct
				switch i {
				case csvConf.Ipublisher:
					csvRecord.publisher = value
				case csvConf.Ititle:
					csvRecord.title = value
				case csvConf.Isbn:
					// clean isbn, remove spaces & dashes
					csvRecord.isbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
					isbnCount++
				case csvConf.Eisbn:
					csvRecord.eisbn = strings.Trim(strings.Replace(value, "-", "", -1), " ")
					eisbnCount++
				case csvConf.Ipubdate:
					csvRecord.pubdate = value
				case csvConf.Iurl:
					csvRecord.url = value
				case csvConf.Ilang:
					csvRecord.lang = value
				}

				// authors are in a slice
				for j := 0; j < len(csvConf.Iauthors); j++ {
					if i == csvConf.Iauthors[j] {
						authors = append(authors, value)
					}
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
	rejectedLinesLog := fmt.Sprintln("rejected lines", rejectedLines)
	logger.Info.Println(rejectedLinesLog)
	userM["rejectedLinesLog"] = rejectedLinesLog

	return csvData, userM, nil
}

// csvSaveProcessed saves cleaned values to a new, clean csv file
// NOTE: this saves persistent data and should thus probably be in models.
func csvSaveProcessed(csvData []CSVRecord, tsname string, userM userMessages) (userMessages, error) {

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
	Isbn := models.Isbn{recordIn.isbn, false}  // print isbn, not electronic
	Eisbn := models.Isbn{recordIn.eisbn, true} // eisbn, electronic
	ebk.Isbns = append(ebk.Isbns, Isbn, Eisbn)
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
