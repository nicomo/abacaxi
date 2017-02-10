package controllers

import (
	"errors"
	"fmt"
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

	// retrieve param passed in url
	vars := mux.Vars(r)
	param := vars["param"]
	matchXML, _ := regexp.MatchString(".xml$", param)
	matchZIP, _ := regexp.MatchString(".zip$", param)

	// hostname from configuration
	conf := config.GetConfig()

	if matchXML {
		ebkID := param[:len(param)-4]

		myEbook, err := models.EbookGetByID(ebkID)
		if err != nil {
			logger.Error.Println(err)
			//TODO: exit cleanly with user message on error
			panic(err)
		}

		// CreateUnimarcFile requires []Ebook
		ebks := make([]models.Ebook, 1)
		ebks = append(ebks, myEbook)

		// create the downloadable file
		fileSize, createFileErr := models.CreateUnimarcFile(ebks, param)
		if createFileErr != nil {
			logger.Error.Println(createFileErr)
		}

		//TODO: abstract in own func (2)
		timeout := time.Duration(5) * time.Second
		transport := &http.Transport{
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get(conf.Hostname + "/static/downloads/" + param)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+param)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		//stream the body, which is a Reader, to the client without fully loading it into memory
		written, err := io.Copy(w, resp.Body)
		if err != nil {
			logger.Error.Println(err)
		}

		// make sure download went OK, then delete file on server
		if fileSize == written {
			// FIXME: url should not be hardcoded
			fDeleteError := os.Remove("./static/downloads/" + param)
			if fDeleteError != nil {
				logger.Error.Println(fDeleteError)
			}
		}

	}

	// download for multiple records
	if matchZIP {
		logger.Debug.Println(matchZIP, param)

		// the name of the target service is in the filename
		tsname := param[:len(param)-4]

		// get the relevant ebooks
		ebks, err := models.EbooksGetWithUnimarcByTSName(tsname)
		if err != nil {
			logger.Error.Println(err)
			//TODO: exit cleanly with user message on error
			panic(err)
		}

		// create the downloadable file
		fileSize, createFileErr := models.CreateUnimarcFile(ebks, tsname+".xml")
		if createFileErr != nil {
			logger.Error.Println(createFileErr)
		}

		// TODO: zip the downloadable file if size too big: > 1*10^6 (i.e. 1Mo)

		//TODO: abstract in own func (2)
		timeout := time.Duration(5) * time.Second
		transport := &http.Transport{
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get(conf.Hostname + "/static/downloads/" + tsname + ".xml")
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+tsname+".xml")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		//stream the body, which is a Reader, to the client without fully loading it into memory
		written, err := io.Copy(w, resp.Body)
		if err != nil {
			logger.Error.Println(err)
		}

		// make sure download went OK, then delete file on server
		if fileSize == written {
			logger.Debug.Printf("filesize: %d - written: %d", fileSize, written)
			// FIXME: url should not be hardcoded
			fDeleteError := os.Remove("./static/downloads/" + tsname + ".xml")
			if fDeleteError != nil {
				logger.Error.Println(fDeleteError)
			}
		}
	}

	if !matchXML && !matchZIP {
		err := errors.New("Couldn't get a known file extension")
		logger.Error.Println(err)
		panic(err)
	}
}
