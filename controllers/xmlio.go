package controllers

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/terryh/goisbn"

	"github.com/gorilla/sessions"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// XMLRecords holds a slice of XMLRecords for parsing
type XMLRecords struct {
	XMLName xml.Name    `xml:"institutional_holdings"`
	Records []XMLRecord `xml:"item"`
}

// XMLRecord single XML record for parsing
type XMLRecord struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	SFXID       string   `xml:"sfx_id"`
	Isbn        string   `xml:"isbn"`
	Eisbn       string   `xml:"eisbn"`
	FirstAuthor string   `xml:"authorlist>author"`
}

// xmlIO takes an xml file to clean it, save copy & unmarshall content
func xmlIO(filename, tsname string, sess *sessions.Session) ([]models.Record, models.TargetService, *sessions.Session, error) {

	// retrieve target service (i.e. ebook package) for this file
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// update date for TS publisher last harvest since
	// we're harvesting books from a publisher provided csv file
	myTS.TSSFXLastHarvest = time.Now()

	// open the source XML file
	file, err := os.Open(filename) // FIXME: should be filepath rather than filename
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	// read the records file
	xmlRecords, err := ReadRecords(file)
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}

	// unmarshall csv records into record structs
	records := []models.Record{}
	for _, record := range xmlRecords {
		record, err := xmlUnmarshall(record, myTS)
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		records = append(records, record)
	}

	// log number of records successfully parsed
	parsedLog := fmt.Sprintf("successfully parsed %d records from %s", len(records), filename)
	logger.Info.Print(parsedLog)
	sess.AddFlash(parsedLog)

	// save a server copy of source xml file
	t := time.Now()
	dst := "./data/" + tsname + "Processed" + t.Format("20060102150405") + ".xml"
	ErrXMLSaveCopy := xmlSaveCopy(dst, filename)
	if ErrXMLSaveCopy != nil {
		logger.Error.Println(ErrXMLSaveCopy)
		return records, myTS, sess, ErrXMLSaveCopy
	}

	// logging + user message with result of save copy
	saveCopyMssg := fmt.Sprintf("successfully saved cleaned up version of xml file as %s", dst)
	logger.Info.Println(saveCopyMssg)

	return records, myTS, sess, nil
}

// ReadRecords reads the XML document
// and returns the array of records nodes
func ReadRecords(reader io.Reader) ([]XMLRecord, error) {
	var xmlRecords XMLRecords
	if err := xml.NewDecoder(reader).Decode(&xmlRecords); err != nil {
		return nil, err
	}
	return xmlRecords.Records, nil
}

// create record object from xml record
func xmlUnmarshall(recordIn XMLRecord, myTS models.TargetService) (models.Record, error) {
	var record models.Record

	record.FirstAuthor = recordIn.FirstAuthor

	// ISBNs : validate & cleanup, convert isbn 10 <-> isbn13
	isbnCleaned, err := goisbn.Cleanup(recordIn.Isbn)
	if err != nil { // not a valid isbn
		return record, err
	}
	IdentifierIsbn := models.Identifier{Identifier: isbnCleaned, IDType: models.IDTypePrint}
	record.Identifiers = append(record.Identifiers, IdentifierIsbn)

	IsbnConverted, err := goisbn.Convert(isbnCleaned)
	if err != nil {
		logger.Error.Printf("couldn't convert isbn: %s - %v", IsbnConverted, err)
	} else {
		IdentifierIsbnConverted := models.Identifier{Identifier: IsbnConverted, IDType: models.IDTypePrint}
		record.Identifiers = append(record.Identifiers, IdentifierIsbnConverted)
	}

	// add sfx ID
	IdentifierSFX := models.Identifier{Identifier: recordIn.SFXID, IDType: models.IDTypeSFX}
	record.Identifiers = append(record.Identifiers, IdentifierSFX)

	record.PublicationTitle = recordIn.Title

	// add target service
	TSEmbed := models.TSEmbed{Name: myTS.TSName, DisplayName: myTS.TSDisplayName}
	record.TargetServices = append(record.TargetServices, TSEmbed)

	if myTS.TSActive {
		record.Active = true
	}

	return record, nil
}

func xmlSaveCopy(dst, src string) error {

	// open the source XML file
	in, err := os.Open(src)
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}
	defer in.Close()

	// create copy file
	out, err := os.Create(dst)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	defer out.Close()

	// do the actual copy
	_, ErrCopy := io.Copy(out, in)
	if ErrCopy != nil {
		logger.Error.Println(ErrCopy)
		return ErrCopy
	}
	ErrClose := out.Close()
	if ErrClose != nil {
		logger.Error.Println(ErrClose)
		return ErrClose
	}

	return nil

}
