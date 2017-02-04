package controllers

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// UploadGetHandler manages upload of a source file
func UploadGetHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be stored in this map
	d := make(map[string]interface{})

	// check if we have messages coming in the Request context
	if userM, ok := fromContextUserM(r.Context()); ok {
		for k, v := range userM {
			d[k] = v
		}
	}

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	views.RenderTmpl(w, "upload", d)

}

// UploadPostHandler receives source file, checks extension
// then passes the file on to the appropriate controller
func UploadPostHandler(w http.ResponseWriter, r *http.Request) {

	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	//FIXME: either userMessages struct of d[string] interface{} as used in other funcs, but not both...
	userM := make(UserMessages)

	// parsing multipart file
	r.ParseMultipartForm(32 << 20)

	// get the Target Service name
	tsname := r.PostFormValue("pack")
	userM["myPackage"] = tsname
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		logger.Error.Println(err)
		return
	}
	defer file.Close()

	// create dir if it doesn't exist
	path := "data"
	pathErr := os.MkdirAll("data", os.ModePerm)
	if pathErr != nil {
		logger.Error.Println(pathErr)
	}

	fpath := path + "/" + handler.Filename
	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	defer f.Close()

	// copy uploaded file into new file
	io.Copy(f, file)

	// if xml pass on to xmlio, if csv, pass on to csvio, if neither, abort
	ext := filepath.Ext(handler.Filename)
	if ext == ".csv" {
		// pass on the name of the package and the name of the file to csvio package

		csvRecords, myTS, userM, err := csvIO(fpath, tsname, userM)
		if err != nil {
			logger.Error.Println(err)
		}
		createdCounter, updatedCounter, createUpdateErr := models.EbooksCreateOrUpdate(csvRecords)
		if createUpdateErr != nil {
			logger.Error.Println("EbooksCreateOrUpdate error: ", createUpdateErr)
		}

		userM["createdCounter"] = strconv.Itoa(createdCounter)
		userM["updatedCounter"] = strconv.Itoa(updatedCounter)

		// update the target service with last update date
		tsUpdateErr := models.TSUpdate(myTS)
		if tsUpdateErr != nil {
			logger.Error.Printf("couldn't update Target Service %v. Error: %v", myTS, tsUpdateErr)
		}

		// list of TS appearing in menu
		TSListing, _ := models.GetTargetServicesListing()
		userM["TSListing"] = TSListing

		views.RenderTmpl(w, "upload-report", userM)

	} else if ext == ".xml" {

		xmlRecords, myTS, userM, err := xmlIO(fpath, tsname, userM)
		if err != nil {
			logger.Error.Println(err)
		}
		createdCounter, updatedCounter, createUpdateErr := models.EbooksCreateOrUpdate(xmlRecords)
		if createUpdateErr != nil {
			logger.Error.Println("EbooksCreateOrUpdate error: ", createUpdateErr)
		}
		userM["createdCounter"] = strconv.Itoa(createdCounter)
		userM["updatedCounter"] = strconv.Itoa(updatedCounter)

		tsUpdateErr := models.TSUpdate(myTS)
		if tsUpdateErr != nil {
			logger.Error.Printf("couldn't update Target Service %v. Error: %v", myTS, tsUpdateErr)
		}

		// list of TS appearing in menu
		TSListing, _ := models.GetTargetServicesListing()
		userM["TSListing"] = TSListing

		views.RenderTmpl(w, "upload-report", userM)

	} else {

		// manage case wrong file extension : message to the user
		logger.Error.Println("wrong file extension")
		userM["wrongExt"] = "wrong file extension"

		// insert the user messages in the http.Request Context before redirecting
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		ctx = newContextUserM(ctx, userM)

		// redirect to upload get page
		UploadGetHandler(w, r.WithContext(ctx))

	}
}
