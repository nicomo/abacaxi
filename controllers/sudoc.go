package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/sudoc"
	"github.com/nicomo/abacaxi/views"
	"github.com/nicomo/gosudoc"
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

	// retrieves the identifiers we care about (isbn/issn)
	var isbn []string
	var issn []string
	var PPN string
	for _, v := range myRecord.Identifiers {
		// we already have a Unimarc record ID
		if v.IDType == models.IDTypePPN {
			//TODO: fetch the unimarc record right away
			PPN = v.Identifier
			break
		}

		// we have Print or online identifiers
		if v.IDType == models.IDTypePrint || v.IDType == models.IDTypeOnline {
			if len(v.Identifier) == 8 { // looks like an ISSN
				issn = append(issn, v.Identifier)
				continue
			}
			isbn = append(isbn, v.Identifier)
		}
	}

	// no PPN yet, let's try and fetch one
	if PPN == "" {
		var res map[string][]string
		// we have an issn
		if len(issn) > 0 && len(isbn) == 0 {
			res, err = gosudoc.Issn2ppn(issn)
			if err != nil {
				logger.Error.Println(err)
			}
		}

		// we have isbns
		res, err = gosudoc.ID2ppn(isbn, "isbn2ppn")
		if err != nil {

			// user friendly message
			msg := fmt.Sprintf("Couldn't find a PPN: %v", err)
			logger.Error.Println(msg)
			sess.AddFlash(msg)
			sess.Save(r, w)

			// redirect & return
			urlStr := "/record/" + recordID
			http.Redirect(w, r, urlStr, http.StatusSeeOther)
			return
		}

		// insert new PPNs into the record struct
		for _, v := range res {
			for _, value := range v {
				var exists bool
				for _, w := range myRecord.Identifiers {
					if value == w.Identifier {
						exists = true
						continue
					}
				}
				if !exists {
					newPPN := models.Identifier{Identifier: value, IDType: models.IDTypePPN}
					myRecord.Identifiers = append(myRecord.Identifiers, newPPN)
					PPN = newPPN.Identifier
				}
			}
		}

		// save the update record struct to DB
		myRecord, err = models.RecordUpdate(myRecord)
		if err != nil {
			logger.Error.Printf("couldn't save PPN to record: %v", err)
		}
	}

	// we have a PPN -> now get the unimarc record
	unimarc, err := sudoc.GetRecord("http://www.sudoc.fr/" + PPN + ".abes")
	if err != nil {
		// user friendly message
		msg := fmt.Sprintf("Couldn't find a Unimarc Record: %v", err)
		logger.Error.Println(msg)
		sess.AddFlash(msg)
		sess.Save(r, w)

		// redirect & return
		urlStr := "/record/" + recordID
		http.Redirect(w, r, urlStr, http.StatusSeeOther)
		return
	}

	// actually save updated ebook struct to DB
	if unimarc != "" {
		myRecord.RecordUnimarc = unimarc
		myRecord, err = models.RecordUpdate(myRecord)
		if err != nil {
			logger.Error.Println(err)
		}
	}

	// user friendly success message
	msg := fmt.Sprintf("Unimarc Record has been saved")
	logger.Debug.Println(msg)
	sess.AddFlash(msg)
	sess.Save(r, w)

	// redirect & return
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, http.StatusSeeOther)
	return

}

// SudocI2PTSHandler retrieves PPNs for all records linked to a Target Service that don't currently have one
func SudocI2PTSHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})
	d["i2pType"] = "Get Sudoc PPN Record IDs for records currently without one"

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myTS"] = tsname

	records, err := models.RecordsGetNoPPNByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
		d["sudoci2pError"] = err
		views.RenderTmpl(w, "sudoci2p-report", d)
	}

	// set up the pipeline
	in := sudoc.GenChannel(records)

	// fan out to 2 workers
	c1 := sudoc.CrawlPPN(in)
	c2 := sudoc.CrawlPPN(in)

	// fan in results
	ppnCounter := 0
	for n := range sudoc.MergeResults(c1, c2) {
		ppnCounter += n
	}

	// let's do a little reporting to the user
	logger.Info.Printf("Number of records : %d - number of records receiving PPNs : %d", len(records), ppnCounter)
	d["RecordsCount"] = len(records)
	d["getPPNResultCount"] = ppnCounter

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "sudoci2p-report", d)
}

// GetRecordsTSHandler retrieves Unimarc Records from Sudoc for all local records using a given target service
func GetRecordsTSHandler(w http.ResponseWriter, r *http.Request) {

	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myTS"] = tsname

	records, err := models.RecordsGetWithPPNByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
		d["sudocGetRecordsError"] = err
		views.RenderTmpl(w, "sudocgetrecords-report", d)
	}

	// set up the pipeline
	in := sudoc.GenChannel(records)

	// fan out to 2 workers
	c1 := sudoc.CrawlRecords(in)
	c2 := sudoc.CrawlRecords(in)

	// fan in results
	recordsCounter := 0
	for n := range sudoc.MergeResults(c1, c2) {
		recordsCounter += n
	}

	// let's do a little reporting to the user
	logger.Info.Printf("Number of local records sent : %d - number of unimarc records received  : %d", len(records), recordsCounter)
	d["RecordsCount"] = len(records)
	d["getRecordsResultCount"] = recordsCounter

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "sudocgetrecords-report", d)
}
