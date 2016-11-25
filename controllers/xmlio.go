package controllers

import (
	"encoding/xml"
	"io"
	"os"

	"github.com/nicomo/minymapp/logger"
)

type XMLRecords struct {
	XMLName xml.Name `xml:"records"`
	Records []XMLRecord
}

type XMLRecord struct {
	XMLName xml.Name `xml:"record"`

	//TODO: map sfx xml to struct
	// see https://www.goinggo.net/2013/06/reading-xml-documents-in-go.html
}

func xmlio(filename string, packname string) {
	logger.Debug.Println(packname)

	// open the XML file
	file, err := os.Open(filename) // FIXME: should be filepath rather than filename
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	// read the records file
	XMLRecords, err := ReadRecords(file)
	if err != nil {
		logger.Error.Println(err)
		os.Exit(1)
	}

	// sanity check, display first record
	logger.Debug.Printf("1st record in file is\n", XMLRecords[0])
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
