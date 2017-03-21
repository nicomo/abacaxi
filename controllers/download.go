package controllers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/config"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
)

// DownloadHandler manages the downloading of marc records:
//   -- test if if the request is either for a single record or a bunch of them
//   -- get the record(s) in DB
//   -- generate the download file & serve it
//   -- delete the downloaded file from server
func DownloadHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve filename passed in url
	vars := mux.Vars(r)
	filename := vars["filename"]
	matchXML, _ := regexp.MatchString(".xml$", filename)
	matchZIP, _ := regexp.MatchString(".zip$", filename)
	var filesize int64

	if !matchXML && !matchZIP {
		err := errors.New("Couldn't get a known file extension")
		logger.Error.Println(err)
		panic(err)
	}

	if matchXML {
		recordID := filename[:len(filename)-4]
		filesize = singleRecordCreateFile(recordID, filename)
	}

	// download for multiple records
	if matchZIP {
		// the name of the target service is in the filename
		tsname := filename[:len(filename)-4]
		filesize, filename = multipleRecordsCreateFile(tsname)
	}

	// hostname from configuration
	conf := config.GetConfig()

	timeout := time.Duration(5) * time.Second
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		DisableKeepAlives:     true,
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(conf.Hostname + "/static/downloads/" + filename)
	if err != nil {
		logger.Error.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	//stream the body, which is a Reader, to the client without fully loading it into memory
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		logger.Error.Println(err)
		panic(err)
	}

	// make sure download went OK, then delete file on server
	if filesize == written {
		ErrFDelete := os.Remove("./static/downloads/" + filename)
		if ErrFDelete != nil {
			logger.Error.Println(ErrFDelete)
		}
	}

}

func multipleRecordsCreateFile(tsname string) (int64, string) {

	// get the relevant records
	records, err := models.RecordsGetWithUnimarcByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
		//TODO: exit cleanly with user message on error
		panic(err)
	}

	// create the downloadable file
	fileSize, ErrCreateFile := models.CreateUnimarcFile(records, tsname+".xml")
	if ErrCreateFile != nil {
		logger.Error.Println(ErrCreateFile)
	}

	// TODO: zip the downloadable file if size too big: > 1*10^6 (i.e. 1Mo)
	// if zipped change the file name
	filename := tsname + ".xml"
	return fileSize, filename

}

func singleRecordCreateFile(recordID string, filename string) int64 {

	// get the relevant record
	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
		//TODO: exit cleanly with user message on error
		panic(err)
	}

	// CreateUnimarcFile requires []Record
	records := make([]models.Record, 1)
	records = append(records, myRecord)

	// create the downloadable file
	fileSize, ErrCreateFile := models.CreateUnimarcFile(records, filename)
	if ErrCreateFile != nil {
		logger.Error.Println(ErrCreateFile)
	}
	return fileSize
}
