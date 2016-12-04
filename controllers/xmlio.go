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

func xmlIO(filename string, packname string, userM userMessages) ([]models.Ebook, userMessages, error) {
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

	// log number of records successfully parsed
	parsedLog := fmt.Sprintf("successfully parsed %d records from %s", len(ebooks), filename)
	logger.Info.Print(parsedLog)
	userM["parsedLog"] = parsedLog

	// save a server copy of source xml file
	t := time.Now()
	dst := "./data/" + packname + "Processed" + t.Format("20060102150405") + ".xml"
	xmlSaveCopyErr := xmlSaveCopy(dst, filename)
	if xmlSaveCopyErr != nil {
		logger.Error.Println(xmlSaveCopyErr)
		return ebooks, userM, xmlSaveCopyErr
	}

	// logging + user message with result of save copy
	saveCopyMssg := fmt.Sprintf("successfully saved cleaned up version of xml file as %s", dst)
	logger.Info.Println(saveCopyMssg)
	userM["saveCopyMssg"] = saveCopyMssg

	return ebooks, userM, nil

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
	ebk.SfxId = recordIn.SfxID

	return ebk
}

func xmlSaveCopy(dst, src string) error {

	// open the source XML file
	in, err := os.Open(src)
	if err != nil {
		logger.Error.Println(err)
		return err
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
