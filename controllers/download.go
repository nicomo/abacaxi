package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
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
	if matchXML {
		ebkID := param[:len(param)-4]

		myEbook, err := models.EbookGetByID(ebkID)
		if err != nil {
			logger.Error.Println(err)
			//TODO: exit cleanly with user message on error
			panic(err)
		}

		fileSize, createFileErr := models.CreateUnimarcFile(myEbook, param)
		if createFileErr != nil {
			logger.Error.Println(createFileErr)
		}

		//TODO: abstract in own func (1)
		timeout := time.Duration(5) * time.Second
		transport := &http.Transport{
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		}

		client := &http.Client{
			Transport: transport,
		}

		// FIXME: url should not be hardcoded
		resp, err := client.Get("http://localhost:8080/static/downloads/" + param)
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

	} else {
		logger.Debug.Println(matchXML, param)
		//TODO : download for multiple records (2)
	}
}
