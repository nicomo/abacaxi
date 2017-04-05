package controllers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

const (
	kbartNumFields = 25
)

func fileIO(filename string, tsname string, userM UserMessages, ext string) ([]models.Record, models.TargetService, UserMessages, error) {
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
	var csvConf map[string]int
	if ext == ".csv" {
		csvConf = csvConfSwap(myTS.TSCsvConf)
		reader.FieldsPerRecord = len(csvConf)
	} else {
		reader.FieldsPerRecord = kbartNumFields // kbart is a const: always 25 fields
	}
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
			rejectedLines = append(rejectedLines, line)
			line++
			continue

		}

		// parse each line into a struct
		record, err := fileParseRow(fRecord, csvConf)
		if err != nil {
			logger.Error.Println(err)
		}

		// add TS to record
		record.TargetServices = append(record.TargetServices, myTS)

		// add record to slice
		records = append(records, record)

		line++
	}

	// log number of records successfully parsed
	parsedLog := fmt.Sprintf("successfully parsed %d lines from %s", len(records), filename)
	logger.Info.Print(parsedLog)
	userM["parsedLog"] = parsedLog

	// log lines rejected
	if len(rejectedLines) > 0 {
		rejectedLinesLog := fmt.Sprintf("rejected lines in file %s: %v", filename, rejectedLines)
		logger.Info.Println(rejectedLinesLog)
		userM["rejectedLinesLog"] = rejectedLinesLog
	}

	return records, myTS, userM, nil
}

func fileParseRow(fRecord []string, csvConf map[string]int) (models.Record, error) {
	var record models.Record

	//TODO: (2) validate row : utf-8 + required fields
	for _, value := range fRecord {
		if !utf8.ValidString(value) { // encoding issue, non utf-8 char. in value
			err := errors.New("parsing failed: non utf-8 char. in value")
			return record, err
		}
	}

	if csvConf == nil { // the csv Configuration is nil, we default to kbart values
		record.PublicationTitle = fRecord[0]

		// ISBNs : validate & cleanup, convert isbn 10 <-> isbn13
		// Identifiers Print ID
		err := getIsbnIdentifiers(fRecord[1], &record, models.IdTypePrint)
		if err != nil && fRecord[1] != "" { // doesn't look like an isbn, might be issn, cleanup and add as is
			idCleaned := strings.Trim(strings.Replace(fRecord[1], "-", "", -1), " ")
			record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IdType: models.IdTypePrint})
		}
		// Identifiers Online ID
		err = getIsbnIdentifiers(fRecord[2], &record, models.IdTypeOnline)
		if err != nil && fRecord[2] != "" { // doesn't look like an isbn, might be issn, cleanup and add as is
			idCleaned := strings.Trim(strings.Replace(fRecord[2], "-", "", -1), " ")
			record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IdType: models.IdTypeOnline})
		}

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
	} else { // we do have a csv configuration
		if i, ok := csvConf["publicationtitle"]; ok {
			record.PublicationTitle = fRecord[i]
		}
		if i, ok := csvConf["identifierprint"]; ok {
			err := getIsbnIdentifiers(fRecord[i], &record, models.IdTypePrint)
			if err != nil && fRecord[i] != "" { // doesn't look like an isbn, might be issn, clean up and add as is
				idCleaned := strings.Trim(strings.Replace(fRecord[i], "-", "", -1), " ")
				record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IdType: models.IdTypePrint})
			}
		}
		if i, ok := csvConf["identifieronline"]; ok {
			err := getIsbnIdentifiers(fRecord[i], &record, models.IdTypeOnline)
			if err != nil && fRecord[i] != "" { // doesn't look like an isbn, might be issn, clean up and add as is
				idCleaned := strings.Trim(strings.Replace(fRecord[i], "-", "", -1), " ")
				record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IdType: models.IdTypeOnline})
			}
		}

		if i, ok := csvConf["datefirstissueonline"]; ok {
			record.DateFirstIssueOnline = fRecord[i]
		}
		if i, ok := csvConf["numfirstvolonline"]; ok {
			record.NumFirstVolOnline = fRecord[i]
		}
		if i, ok := csvConf["numfirstissueonline"]; ok {
			record.NumFirstIssueOnline = fRecord[i]
		}
		if i, ok := csvConf["datelastissueonline"]; ok {
			record.DateLastIssueOnline = fRecord[i]
		}
		if i, ok := csvConf["numlastvolonline"]; ok {
			record.NumLastVolOnline = fRecord[i]
		}
		if i, ok := csvConf["numlastissueonline"]; ok {
			record.NumLastIssueOnline = fRecord[i]
		}
		if i, ok := csvConf["titleurl"]; ok {
			record.TitleURL = fRecord[i]
		}
		if i, ok := csvConf["firstauthor"]; ok {
			record.FirstAuthor = fRecord[i]
		}
		if i, ok := csvConf["titleid"]; ok {
			record.TitleID = fRecord[i]
		}
		if i, ok := csvConf["embargoinfo"]; ok {
			record.EmbargoInfo = fRecord[i]
		}
		if i, ok := csvConf["coveragedepth"]; ok {
			record.CoverageDepth = fRecord[i]
		}
		if i, ok := csvConf["notes"]; ok {
			record.Notes = fRecord[i]
		}
		if i, ok := csvConf["publishername"]; ok {
			record.PublisherName = fRecord[i]
		}
		if i, ok := csvConf["publicationtype"]; ok {
			record.PublicationType = fRecord[i]
		}
		if i, ok := csvConf["datemonographpublishedprint"]; ok {
			record.DateMonographPublishedPrint = fRecord[i]
		}
		if i, ok := csvConf["datemonographpublishedonline"]; ok {
			record.DateMonographPublishedOnline = fRecord[i]
		}
		if i, ok := csvConf["monographvolume"]; ok {
			record.MonographVolume = fRecord[i]
		}
		if i, ok := csvConf["monographedition"]; ok {
			record.MonographEdition = fRecord[i]
		}
		if i, ok := csvConf["firsteditor"]; ok {
			record.FirstEditor = fRecord[i]
		}
		if i, ok := csvConf["parentpublicationtitleid"]; ok {
			record.ParentPublicationTitleID = fRecord[i]
		}
		if i, ok := csvConf["precedingpublicationtitleid"]; ok {
			record.PrecedingPublicationTitleID = fRecord[i]
		}
		if i, ok := csvConf["accesstype"]; ok {
			record.AccessType = fRecord[i]
		}
	}

	if !validateRecord(record) {
		recordNotValid := errors.New("record not valid")
		return record, recordNotValid
	}

	record.DateCreated = time.Now()

	return record, nil
}

func validateRecord(record models.Record) bool {
	if len(record.Identifiers) == 0 || record.PublicationTitle == "" {
		return false
	}
	return true
}
