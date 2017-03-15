package controllers

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

func fileIO(filename string, tsname string, ext string) ([]models.Record, models.TargetService, error) {
	// retrieve target service (i.e. ebook package) for this file
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}
	// update date for TS publisher last harvest since
	// we're harvesting books from a publisher provided csv file
	myTS.TSPublisherLastHarvest = time.Now()

	// open csv file
	csvFile, err := os.Open(filename)
	if err != nil {
		logger.Error.Println("cannot open csv file")
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// slice will hold successfully parsed records
	var records []models.Record

	// package csv has n fields, separator is ;
	// TODO: (1) retrieve the number of fields in csvConf if it's a .csv file, or else constant if it's .kbart
	reader.FieldsPerRecord = 25
	reader.Comma = ';'

	// counters to keep track of records parsed, for logging
	line := 1
	var rejectedLines []int

	for {
		// read a row
		fRecord, err := reader.Read()

		// if at EOF, break out of loop
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error.Println(err)
			panic(err)
		}

		// if row not of the expected length, move on
		if len(fRecord) != reader.FieldsPerRecord {
			logger.Info.Printf("parsing line %d failed: invalid length of %d, expected %d\n", line, len(fRecord), reader.FieldsPerRecord)
			rejectedLines = append(rejectedLines, line)
			continue
		}

		// parse each line into a struct
		record, err := fileParseRow(fRecord, ext)
		if err != nil {
			logger.Error.Println(err)
		}

		// add TS to record
		record.TargetServices = append(record.TargetServices, myTS)

		// add record to slice
		records = append(records, record)

		line++
	}

	//TODO: log number of records successfully parsed
	/*	parsedLog := fmt.Sprintf("successfully parsed %d lines from %s", len(records), filename)
		logger.Info.Print(parsedLog)
		userM["parsedLog"] = parsedLog

		// log lines rejected
		if len(rejectedLines) > 0 {
			rejectedLinesLog := fmt.Sprintf("rejected lines in file %s: %v", filename, rejectedLines)
			logger.Info.Println(rejectedLinesLog)
			userM["rejectedLinesLog"] = rejectedLinesLog
		}
	*/
	return records, myTS, nil
}

func fileParseRow(fRecord []string, ext string) (models.Record, error) {

	var record models.Record
	var identifiers []models.Identifier

	//TODO: (2) validate row : utf-8 + required fields
	for _, value := range fRecord {
		if !utf8.ValidString(value) { // encoding issue, non utf-8 char. in value
			err := errors.New("parsing failed: non utf-8 char. in value")
			return record, err
		}
	}

	if ext == ".kbart" {
		record.PublicationTitle = fRecord[0]
		// Identifiers Print ID
		printID := strings.Trim(strings.Replace(fRecord[1], "-", "", -1), " ")
		identifiers = append(identifiers, models.Identifier{Identifier: printID, IdType: models.IdTypePrint})
		// Identifiers Online ID
		onlineID := strings.Trim(strings.Replace(fRecord[2], "-", "", -1), " ")
		identifiers = append(identifiers, models.Identifier{Identifier: onlineID, IdType: models.IdTypeOnline})
		record.Identifiers = identifiers
		record.DateFirstIssueOnline = fRecord[3]
		record.NumFirstVolOnline = fRecord[4]
		record.NumFirstIssueOnline = fRecord[5]
		record.DateLastIssueOnline = fRecord[6]
		record.NumLastVolOnline = fRecord[7]
		record.NumLastIssueOnline = fRecord[8]
		record.TitleURL = fRecord[9]
		record.FirstAuthor = fRecord[10]
		record.TitleID = fRecord[11]
		record.EmbargoInfo = fRecord[12]
		record.CoverageDepth = fRecord[13]
		record.Notes = fRecord[14]
		record.PublisherName = fRecord[15]
		record.PublicationType = fRecord[16]
		record.DateMonographPublishedPrint = fRecord[17]
		record.DateMonographPublishedOnline = fRecord[18]
		record.MonographVolume = fRecord[19]
		record.MonographEdition = fRecord[20]
		record.FirstEditor = fRecord[21]
		record.ParentPublicationTitleID = fRecord[22]
		record.PrecedingPublicationTitleID = fRecord[23]
		record.AccessType = fRecord[24]
	}

	if !validateRecord(record) {
		recordNotValid := errors.New("record not valid")
		return record, recordNotValid
	}

	return record, nil
}

func validateRecord(record models.Record) bool {
	if len(record.Identifiers) == 0 || record.PublicationTitle == "" {
		return false
	}
	return true
}
