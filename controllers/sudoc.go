package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/sudoc"
)

// GetSudocRecordHandler takes Identifiers (e.g. ISBN) and asks the Sudoc Web Service for a Unimarc record
func GetSudocRecordHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	// retrieve the record
	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
		// redirect
		redirectURL := "/record/" + recordID
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	if err := sudoc.GetSudocRecord(myRecord); err != nil {
		// user friendly error message
		msg := fmt.Sprintf("Unimarc Record couldn't be retrieve: %v", err)
		sess.AddFlash(msg)
		sess.Save(r, w)

		// redirect & return
		urlStr := "/record/" + recordID
		http.Redirect(w, r, urlStr, http.StatusSeeOther)
		return
	}

	// user friendly success message
	msg := fmt.Sprintf("Unimarc Record has been saved")
	sess.AddFlash(msg)
	sess.Save(r, w)

	// redirect & return
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, http.StatusSeeOther)
	return

}

// GetSudocRecordsHandler retrieves Unimarc Records from Sudoc for all local records using a given target service
func GetSudocRecordsHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]

	records, err := models.RecordsGetNoPPNByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// we have records, and can proceed
	// - redirect user to home with a flash message
	// - continue our work in a separate go routine
	sess.AddFlash("Request is running in the background, result will be in the reports")
	sess.Save(r, w)
	go sudoc.GetSudocRecords(records)
	http.Redirect(w, r, "/", http.StatusFound)

}
