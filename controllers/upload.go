package controllers

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" { // you just arrived here, I'll give you a token
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		// parsing multipart file
		r.ParseMultipartForm(32 << 20)
		// get the ebook package name
		packname := r.PostFormValue("pack")
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

		// TODO: if xml pass on to xmlio, if csv, pass on to csvio

		// pass on the name of the package and the name of the file to csvio package
		csvRecords, err := csvIO(path, packname)
		if err != nil {
			logger.Error.Println(err)
		}

		models.EbooksCreateOrUpdate(csvRecords)
	}
}
