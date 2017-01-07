package controllers

import (
	"encoding/xml"
	"fmt"
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
}

// xmlIO takes an xml file to clean it, save copy & unmarshall content
func xmlIO(filename string, tsname string, userM userMessages) ([]models.Ebook, models.TargetService, userMessages, error) {
	logger.Debug.Println(tsname)

	// retrieve target service (i.e. ebook package) for this file
	myTargetService, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}
	logger.Debug.Println(myTargetService)
	// update date for TS publisher last harvest since
	// we're harvesting books from a publisher provided csv file
	myTargetService.TSSFXLastHarvest = time.Now()

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
		ebook := xmlUnmarshall(record, myTargetService)
		ebooks = append(ebooks, ebook)
	}

	// log number of records successfully parsed
	parsedLog := fmt.Sprintf("successfully parsed %d records from %s", len(ebooks), filename)
	logger.Info.Print(parsedLog)
	userM["parsedLog"] = parsedLog

	// save a server copy of source xml file
	t := time.Now()
	dst := "./data/" + tsname + "Processed" + t.Format("20060102150405") + ".xml"
	xmlSaveCopyErr := xmlSaveCopy(dst, filename)
	if xmlSaveCopyErr != nil {
		logger.Error.Println(xmlSaveCopyErr)
		return ebooks, myTargetService, userM, xmlSaveCopyErr
	}

	// logging + user message with result of save copy
	saveCopyMssg := fmt.Sprintf("successfully saved cleaned up version of xml file as %s", dst)
	logger.Info.Println(saveCopyMssg)
	userM["saveCopyMssg"] = saveCopyMssg

	return ebooks, myTargetService, userM, nil
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
func xmlUnmarshall(recordIn XMLRecord, myTargetService models.TargetService) models.Ebook {
	ebk := models.Ebook{}
	for _, aut := range recordIn.Authors {
		ebk.Authors = append(ebk.Authors, aut)
	}
	Isbn := models.Isbn{strings.Trim(strings.Replace(recordIn.Isbn, "-", "", -1), " "), false, false} // print isbn, not electronic, not primary
	Eisbn := models.Isbn{strings.Trim(strings.Replace(recordIn.Eisbn, "-", "", -1), " "), true, true} // eisbn, electronic, primary
	ebk.Isbns = append(ebk.Isbns, Isbn, Eisbn)
	ebk.Title = recordIn.Title
	ebk.SFXLastHarvest = time.Now()
	ebk.TargetService = append(ebk.TargetService, myTargetService)
	ebk.SfxId = recordIn.SfxID

	return ebk
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
	_, copyErr := io.Copy(out, in)
	if copyErr != nil {
		logger.Error.Println(copyErr)
		return copyErr
	}
	closeErr := out.Close()
	if closeErr != nil {
		logger.Error.Println(closeErr)
		return closeErr
	}

	return nil

}
