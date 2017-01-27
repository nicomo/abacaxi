package controllers

import (
	"fmt"
	"io"
	"net/http"
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
		logger.Debug.Println(matchXML, param)
		ebkID := param[:len(param)-4]
		logger.Debug.Println(ebkID)

		myEbook, err := models.EbookGetByID(ebkID)
		if err != nil {
			logger.Error.Println(err)
			//TODO: exit cleanly with user message on error
			panic(err)
		}

		createFileErr := models.CreateUnimarcFile(myEbook, param)
		if createFileErr != nil {
			logger.Error.Println(createFileErr)
		}

		timeout := time.Duration(5) * time.Second
		transport := &http.Transport{
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://localhost:8080/static/downloads/" + param)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+param)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		//stream the body to the client without fully loading it into memory
		io.Copy(w, resp.Body)

	} else {
		logger.Debug.Println(matchXML, param)
	}
}
