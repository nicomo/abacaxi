package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

type parseparams struct {
	tsname    string
	fpath     string
	filetype  string
	delimiter rune
	csvconf   map[string]int
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

	// get the file delimiter (for csv and kbart), defaulting to tab
	delimiter := rune('\t')
	if r.PostFormValue("delimiter") == "semicolon" {
		delimiter = ';'
	}

	// get the optional csv fields
	csvconf, err := getCSVParams(r)
	if filetype == "publishercsv" && err != nil {
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
	fpath := path + "/" + time.Now().Format("2006-01-02-15:04:05") + "-" + handler.Filename
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
		delimiter,
		csvconf,
	}

	// we have a file to parse
	// let's do that in a separate go routine
	go parseFile(pp)

	// and redirect the user home with a flash message
	sess.AddFlash("Upload is running in the background, result will be in the reports")
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func parseFile(pp parseparams) {
	var (
		records []models.Record
		report  models.Report
		err     error
	)

	if pp.filetype == "sfxxml" {
		records, err = xmlIO(pp, &report)
		if err != nil {
			logger.Error.Println(err)
			report.Success = false
			report.Text = append(report.Text, fmt.Sprintf("Upload process couldn't complete: %v", err))
			report.ReportCreate()
			return
		}
		report.ReportType = models.UploadSfx
	} else if pp.filetype == "publishercsv" || pp.filetype == "kbart" {
		records, err = fileIO(pp, &report)
		if err != nil {
			logger.Error.Println(err)
			report.Success = false
			report.Text = append(report.Text, fmt.Sprintf("Upload process couldn't complete: %v", err))
			report.ReportCreate()
			return
		}
		if pp.filetype == "publishercsv" {
			report.ReportType = models.UploadCsv
		}
		if pp.filetype == "kbart" {
			report.ReportType = models.UploadKbart
		}
	} else {
		// manage case wrong file extension : message to the user
		logger.Error.Println("unknown file type")
		report.Success = false
		report.Text = append(report.Text, fmt.Sprintln("unknown file type"))
		report.ReportCreate()
		return
	}

	// save the records to DB
	recordsUpdated, recordsInserted := models.RecordsUpsert(records)

	// report
	report.Text = append(report.Text, fmt.Sprintf("Updated %d records / Inserted %d records",
		recordsUpdated,
		recordsInserted))
	report.Success = true

	// save the report to DB
	if err := report.ReportCreate(); err != nil {
		logger.Error.Printf("couldn't save the report to DB: %v", err)
	}

}
