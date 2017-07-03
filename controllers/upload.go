package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

type parseparams struct {
	tsname   string
	fpath    string
	filetype string
	csvconf  map[string]int
}

// UploadGetHandler manages upload of a source file
func UploadGetHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be stored in this map
	d := make(map[string]interface{})

	// Get session
	sess := session.Instance(r)
	if sess.Values["id"] != nil {
		d["IsLoggedIn"] = true
	}

	// Get flash messages, if any.
	if flashes := sess.Flashes(); len(flashes) > 0 {
		d["Flashes"] = flashes
	}
	sess.Save(r, w)

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	views.RenderTmpl(w, "upload", d)

}

// UploadPostHandler receives source file, checks extension
// then passes the file on to the appropriate controller
func UploadPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get session, to be used for feedback flash messages
	sess := session.Instance(r)

	// parsing multipart file
	r.ParseMultipartForm(32 << 20)

	// get the Target Service name and the file type
	tsname := r.PostFormValue("tsname")
	filetype := r.PostFormValue("filetype")
	logger.Debug.Printf("++++++++ FILETYPE : %s", filetype)

	// get the optional csv fields
	csvconf, err := getCSVParams(r)
	if err != nil {
		logger.Error.Printf("couldn't get csv params: %v", err)
	}

	// upload the file
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		logger.Error.Println(err)
		return
	}
	defer file.Close()

	// create dir if it doesn't exist
	path := "data"
	ErrPath := os.MkdirAll("data", os.ModePerm)
	if ErrPath != nil {
		logger.Error.Println(ErrPath)
	}

	// open newly created file
	fpath := path + "/" + handler.Filename
	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	defer f.Close()

	// copy uploaded file into new file
	io.Copy(f, file)

	pp := parseparams{
		tsname,
		fpath,
		filetype,
		csvconf,
	}
	logger.Debug.Printf("parseparams: %v", pp)

	var records []models.Record
	var report string
	if filetype == "sfxxml" {
		records, report, err = xmlIO(pp)
		if err != nil {
			logger.Error.Println(err)
			sess.AddFlash(err)
			sess.Save(r, w)
			// redirect to upload get page
			UploadGetHandler(w, r)
			return
		}
	} else if filetype == "publishercsv" || filetype == "kbart" {
		records, report, err = fileIO(pp)
		if err != nil {
			logger.Error.Println(err)
			sess.AddFlash(err)
			sess.Save(r, w)
			// redirect to upload get page
			UploadGetHandler(w, r)
			return
		}
	} else {
		// manage case wrong file extension : message to the user
		logger.Error.Println("wrong file extension")
		sess.AddFlash("could not recognize the file extension")
		sess.Save(r, w)
		// redirect to upload get page
		UploadGetHandler(w, r)
		return
	}

	recordsUpdated, recordsInserted := models.RecordsUpsert(records)

	reportCount := fmt.Sprintf("Updated %d records / Inserted %d records",
		recordsUpdated,
		recordsInserted)
	report += reportCount
	sess.AddFlash(report)
	sess.Save(r, w)

	redirectURL := "/ts/display/" + tsname
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
