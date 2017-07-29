package controllers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

// RecordHandler displays a single record
func RecordHandler(w http.ResponseWriter, r *http.Request) {
	// data to be display in UI will be stored in this map
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

	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	// format the dates
	d["formattedDateCreated"] = myRecord.DateCreated.Format(time.RFC822)

	if !myRecord.DateUpdated.IsZero() {
		d["formattedDateUpdated"] = myRecord.DateUpdated.Format(time.RFC822)
	}

	d["Record"] = myRecord

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "record", d)
}

// RecordDeleteHandler handles deleting a single ebook
func RecordDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	err := models.RecordDelete(recordID)
	if err != nil {
		logger.Error.Println(err)

		// TODO: transmit either error or success message to user

		// redirect
		redirectURL := "/record/" + recordID
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}

	// redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RecordExportUnimarcHandler exports a single unimarc record
// To export a batch of records, see targetservice.go
func RecordExportUnimarcHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve record ID
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	// get the relevant record
	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
		//TODO: exit cleanly with user message on error
		panic(err)
	}

	// put record in slice (required by models.CreateUnimarcFile)
	recordToExport := []models.Record{myRecord}
	filename := recordID + ".xml"

	// create the file
	filesize, err := models.CreateUnimarcFile(recordToExport, filename)
	if err != nil {
		logger.Error.Printf("could not create file: %v", err)
		//TODO: exit cleanly with user message on error
	}

	// export the file
	if err := exportFile(w, r, filename, filesize); err != nil {
		logger.Error.Printf("couldn't stream the export file: %v", err)
	}

}

//RecordToggleAcquiredHandler toggles the boolean value "acquired" for a record
func RecordToggleAcquiredHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	if myRecord.Acquired {
		myRecord.Acquired = false
	} else {
		myRecord.Acquired = true
	}

	myRecord, err = models.RecordUpdate(myRecord)
	if err != nil {
		logger.Error.Println(err)
	}

	// refresh record page
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, http.StatusSeeOther)
}

// RecordToggleActiveHandler toggles the boolean value "active" for an record
func RecordToggleActiveHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	if myRecord.Active {
		myRecord.Active = false
	} else {
		myRecord.Active = true
	}

	_, err = models.RecordUpdate(myRecord)
	if err != nil {
		logger.Error.Println(err)
	}

	// refresh record page
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, http.StatusSeeOther)
}
