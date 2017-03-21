package controllers

import (
	"net/http"
	"time"

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

	// record ID is last part of the URL
	recordID := r.URL.Path[len("/record/"):]

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
	// record ID is last part of the URL
	recordID := r.URL.Path[len("/record/delete/"):]

	err := models.RecordDelete(recordID)
	if err != nil {
		logger.Error.Println(err)

		// TODO: transmit either error or success message to user

		// redirect to home
		redirectURL := "/record/" + recordID
		http.Redirect(w, r, redirectURL, 303)
	}

	// redirect to home
	http.Redirect(w, r, "/", 303)
}

//RecordToggleAcquiredHandler toggles the boolean value "acquired" for a record
func RecordToggleAcquiredHandler(w http.ResponseWriter, r *http.Request) {

	// record ID is last part of the URL
	recordID := r.URL.Path[len("/record/toggleacquired/"):]

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
	http.Redirect(w, r, urlStr, 303)
}

// RecordToggleActiveHandler toggles the boolean value "active" for an record
func RecordToggleActiveHandler(w http.ResponseWriter, r *http.Request) {

	// record ID is last part of the URL
	recordID := r.URL.Path[len("/record/toggleactive/"):]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	if myRecord.Active {
		myRecord.Active = false
	} else {
		myRecord.Active = true
	}

	myRecord, err = models.RecordUpdate(myRecord)
	if err != nil {
		logger.Error.Println(err)
	}

	// refresh record page
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, 303)
}
