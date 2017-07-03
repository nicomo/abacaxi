package controllers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

const (
	kbartNumFields = 25
)

func fileIO(pp parseparams) ([]models.Record, string, error) {
	// slice will hold successfully parsed records
	var records []models.Record

	// retrieve target service (e.g. ebook package) for this file
	myTS, err := models.GetTargetService(pp.tsname)
	if err != nil {
		return records, "", err
	}

	// open file
	f, err := os.Open(pp.fpath)
	if err != nil {
		return nil, "", errors.New("cannot open the source file")
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// target service csv has n fields, separator is ;
	if pp.filetype == "publishercsv" {
		reader.FieldsPerRecord = len(pp.csvconf)
	} else {
		reader.FieldsPerRecord = kbartNumFields // kbart is a const: always 25 fields
	}
	reader.Comma = ';'

	// counters to keep track of records parsed, for logging
	line := 1
	var rejectedLines []int

	for {
		// read a row
		r, err := reader.Read()

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

		// validate we don't have an unicode replacement char. in the string
		// if we do abort: source file isn't proper utf8
		for _, v := range r {
			if strings.ContainsRune(v, '\uFFFD') {
				err := errors.New("parsing failed: non utf-8 character in file")
				return records, "", err
			}
		}

		// parse each line into a struct
		record, err := fileParseRow(r, pp.csvconf)
		if err != nil {
			logger.Error.Println(err, r)
			continue
		}

		// add TS to record
		record.TargetServices = append(record.TargetServices, myTS)

		// add record to slice
		records = append(records, record)

		line++
	}

	// log number of records successfully parsed
	report := fmt.Sprintf(
		`successfully parsed %d lines from %s
		lines rejected in source file: %v`,
		len(records), pp.fpath, rejectedLines)

	// no records parsed
	if len(records) == 0 {
		err := errors.New("couldn't parse a single line: check your input file")
		return records, "", err

	}

	return records, report, nil
}

func fileParseRow(row []string, csvConf map[string]int) (models.Record, error) {

	var record models.Record

	if csvConf == nil { // the csv Configuration is nil, we default to kbart values

		record.PublicationTitle = row[0]

		// ISBNs : validate & cleanup, convert isbn 10 <-> isbn13
		// Identifiers Print ID
		err := getIsbnIdentifiers(row[1], &record, models.IDTypePrint)
		if err != nil && row[1] != "" { // doesn't look like an isbn, might be issn, cleanup and add as is
			idCleaned := strings.Trim(strings.Replace(row[1], "-", "", -1), " ")
			record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IDType: models.IDTypePrint})
		}
		// Identifiers Online ID
		err = getIsbnIdentifiers(row[2], &record, models.IDTypeOnline)
		if err != nil && row[2] != "" { // doesn't look like an isbn, might be issn, cleanup and add as is
			idCleaned := strings.Trim(strings.Replace(row[2], "-", "", -1), " ")
			record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IDType: models.IDTypeOnline})
		}

		record.DateFirstIssueOnline = row[3]
		record.NumFirstVolOnline = row[4]
		record.NumFirstIssueOnline = row[5]
		record.DateLastIssueOnline = row[6]
		record.NumLastVolOnline = row[7]
		record.NumLastIssueOnline = row[8]
		record.TitleURL = row[9]
		record.FirstAuthor = row[10]
		record.TitleID = row[11]
		record.EmbargoInfo = row[12]
		record.CoverageDepth = row[13]
		record.Notes = row[14]
		record.PublisherName = row[15]
		record.PublicationType = row[16]
		record.DateMonographPublishedPrint = row[17]
		record.DateMonographPublishedOnline = row[18]
		record.MonographVolume = row[19]
		record.MonographEdition = row[20]
		record.FirstEditor = row[21]
		record.ParentPublicationTitleID = row[22]
		record.PrecedingPublicationTitleID = row[23]
		record.AccessType = row[24]
	} else { // we do have a csv configuration
		if i, ok := csvConf["publicationtitle"]; ok {
			record.PublicationTitle = row[i-1]
		}
		if i, ok := csvConf["identifierprint"]; ok {
			err := getIsbnIdentifiers(row[i-1], &record, models.IDTypePrint)
			if err != nil && row[i-1] != "" { // doesn't look like an isbn, might be issn, clean up and add as is
				idCleaned := strings.Trim(strings.Replace(row[i-1], "-", "", -1), " ")
				record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IDType: models.IDTypePrint})
			}
		}
		if i, ok := csvConf["identifieronline"]; ok {
			err := getIsbnIdentifiers(row[i-1], &record, models.IDTypeOnline)
			if err != nil && row[i-1] != "" { // doesn't look like an isbn, might be issn, clean up and add as is
				idCleaned := strings.Trim(strings.Replace(row[i-1], "-", "", -1), " ")
				record.Identifiers = append(record.Identifiers, models.Identifier{Identifier: idCleaned, IDType: models.IDTypeOnline})
			}
		}

		if i, ok := csvConf["datefirstissueonline"]; ok {
			record.DateFirstIssueOnline = row[i-1]
		}
		if i, ok := csvConf["numfirstvolonline"]; ok {
			record.NumFirstVolOnline = row[i-1]
		}
		if i, ok := csvConf["numfirstissueonline"]; ok {
			record.NumFirstIssueOnline = row[i-1]
		}
		if i, ok := csvConf["datelastissueonline"]; ok {
			record.DateLastIssueOnline = row[i-1]
		}
		if i, ok := csvConf["numlastvolonline"]; ok {
			record.NumLastVolOnline = row[i-1]
		}
		if i, ok := csvConf["numlastissueonline"]; ok {
			record.NumLastIssueOnline = row[i-1]
		}
		if i, ok := csvConf["titleurl"]; ok {
			record.TitleURL = row[i-1]
		}
		if i, ok := csvConf["firstauthor"]; ok {
			record.FirstAuthor = row[i-1]
		}
		if i, ok := csvConf["titleid"]; ok {
			record.TitleID = row[i-1]
		}
		if i, ok := csvConf["embargoinfo"]; ok {
			record.EmbargoInfo = row[i-1]
		}
		if i, ok := csvConf["coveragedepth"]; ok {
			record.CoverageDepth = row[i-1]
		}
		if i, ok := csvConf["notes"]; ok {
			record.Notes = row[i-1]
		}
		if i, ok := csvConf["publishername"]; ok {
			record.PublisherName = row[i-1]
		}
		if i, ok := csvConf["publicationtype"]; ok {
			record.PublicationType = row[i-1]
		}
		if i, ok := csvConf["datemonographpublishedprint"]; ok {
			record.DateMonographPublishedPrint = row[i-1]
		}
		if i, ok := csvConf["datemonographpublishedonline"]; ok {
			record.DateMonographPublishedOnline = row[i-1]
		}
		if i, ok := csvConf["monographvolume"]; ok {
			record.MonographVolume = row[i-1]
		}
		if i, ok := csvConf["monographedition"]; ok {
			record.MonographEdition = row[i-1]
		}
		if i, ok := csvConf["firsteditor"]; ok {
			record.FirstEditor = row[i-1]
		}
		if i, ok := csvConf["parentpublicationtitleid"]; ok {
			record.ParentPublicationTitleID = row[i-1]
		}
		if i, ok := csvConf["precedingpublicationtitleid"]; ok {
			record.PrecedingPublicationTitleID = row[i-1]
		}
		if i, ok := csvConf["accesstype"]; ok {
			record.AccessType = row[i-1]
		}
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
