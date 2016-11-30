package controllers

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/views"
)

// UploadHandler manages upload of source file, checks extension
// then passes the file on to the appropriate controller
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" { // you just arrived here, I'll give you a token
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {

		// our messages (errors, confirmation, etc) to the user & the template will be store in this map
		userM := make(userMessages)

		// parsing multipart file
		r.ParseMultipartForm(32 << 20)

		// get the ebook package name
		packname := r.PostFormValue("pack")
		logger.Debug.Println(packname)
		userM["packname"] = packname
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			logger.Error.Println(err)
			return
		}
		defer file.Close()

		// create new file with same name
		path := "./data/" + handler.Filename
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		defer f.Close()

		// copy uploaded file into new file
		io.Copy(f, file)

		// if xml pass on to xmlio, if csv, pass on to csvio, if neither, abort
		ext := filepath.Ext(handler.Filename)
		fmt.Println(ext)
		if ext == ".csv" {
			// pass on the name of the package and the name of the file to csvio package

			csvRecords, userM, err := csvIO(path, packname, userM)
			if err != nil {
				logger.Error.Println(err)
			}
			createUpdateErr := models.EbooksCreateOrUpdate(csvRecords)
			if createUpdateErr != nil {
				logger.Error.Println("EbooksCreateOrUpdate error: ", createUpdateErr)
			}
			views.RenderTmpl(w, "upload", userM)

		} else if ext == ".xml" {
			xmlRecords, userM, err := xmlIO(path, packname, userM)
			if err != nil {
				logger.Error.Println(err)
			}
			createUpdateErr := models.EbooksCreateOrUpdate(xmlRecords)
			if createUpdateErr != nil {
				logger.Error.Println("EbooksCreateOrUpdate error: ", createUpdateErr)
			}
			views.RenderTmpl(w, "upload", userM)
		} else {
			logger.Debug.Println("wrong file extension")
		}

	}
}
