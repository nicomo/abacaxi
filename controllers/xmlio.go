package controllers

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
)

type XMLRecords struct {
	XMLName xml.Name    `xml:"institutional_holdings"`
	Records []XMLRecord `xml:"item"`
}

type XMLRecord struct {
	XMLName xml.Name `xml:"item"`
	Title   string   `xml:"title"`
	SfxID   string   `xml:"sfx_id"`
	Isbn    string   `xml:"isbn"`
	Eisbn   string   `xml:"eisbn"`
	Authors []string `xml:"authorlist>author"`

	//TODO: map sfx xml to struct
	// see https://www.goinggo.net/2013/06/reading-xml-documents-in-go.html
}

func xmlIO(filename string, packname string) ([]models.Ebook, error) {
	logger.Debug.Println(packname)

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

	// sanity check, display first record
	logger.Debug.Println("1st record in file is\n", xmlRecords[0])

	// unmarshall csv records into ebook structs
	ebooks := []models.Ebook{}
	for _, record := range xmlRecords {
		ebook := xmlUnmarshall(record, packname)
		ebooks = append(ebooks, ebook)
	}

	// save a server copy of source xml file
	t := time.Now()
	dst := "./data/cairn_" + t.Format("20060102150405") + ".xml"
	xmlSaveCopy(dst, filename)

	return ebooks, nil

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

// create ebook object from xml record
func xmlUnmarshall(recordIn XMLRecord, packname string) models.Ebook {
	ebk := models.Ebook{}
	for _, aut := range recordIn.Authors {
		ebk.Authors = append(ebk.Authors, aut)
	}
	Isbn := models.Isbn{strings.Trim(strings.Replace(recordIn.Isbn, "-", "", -1), " "), false, false} // print isbn, not electronic, not primary
	Eisbn := models.Isbn{strings.Trim(strings.Replace(recordIn.Eisbn, "-", "", -1), " "), true, true} // eisbn, electronic, primary
	ebk.Isbns = append(ebk.Isbns, Isbn, Eisbn)
	ebk.Title = recordIn.Title
	ebk.SFXLastHarvest = time.Now()
	ebk.TargetService = packname

	return ebk
}

func xmlSaveCopy(dst, src string) {

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
	}
	defer out.Close()

	// do the actual copy
	_, copyErr := io.Copy(out, in)
	if copyErr != nil {
		logger.Error.Println(copyErr)
	}
	closeErr := out.Close()
	if closeErr != nil {
		logger.Error.Println(closeErr)
	}

}
