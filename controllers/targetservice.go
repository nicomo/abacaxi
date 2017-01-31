package controllers

import (
	"net/http"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// TargetServiceHandler retrieves the ebooks linked to a Target Service
//  and various other info, e.g. number of library records linked, etc.
func TargetServiceHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// package name is last part of the URL
	tsname := r.URL.Path[len("/package/"):]
	d["myPackage"] = tsname

	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}
	d["IsTSActive"] = myTS.TSActive

	count := models.TSCountEbooks(tsname)
	d["myPackageEbooksCount"] = count

	if count > 0 { // no need to query for actual ebooks otherwise

		// how many ebooks have marc records
		nbRecordsUnimarc := models.TSCountRecordsUnimarc(tsname)
		d["myPackageRecordsUnimarcCount"] = nbRecordsUnimarc

		// how many ebooks have a PPN from the Sudoc Union Catalog
		nbPPNs := models.TSCountPPNs(tsname)
		d["myPackagePPNsCount"] = nbPPNs

		// get the ebooks
		records, err := models.EbooksGetByTSName(tsname)
		if err != nil {
			logger.Error.Println(err)
		}
		d["myRecords"] = records
	}

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "targetservice", d)
}

// TargetServiceNewGetHandler displays the form to register a new Target Service (i.e. ebook package)
func TargetServiceNewGetHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	views.RenderTmpl(w, "targetservicenewget", d)
}

// TargetServiceNewPostHandler manages the form to register a new Target Service (i.e. ebook package)
func TargetServiceNewPostHandler(w http.ResponseWriter, r *http.Request) {
	d := make(map[string]interface{})

	err := models.TSCreate(r)
	if err != nil {
		d["tsCreateErr"] = err
		logger.Error.Println(err)
		views.RenderTmpl(w, "targetservicenewget", d)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// TargetServiceToggleActiveHandler changes the boolean "active" for a TS *and* records who are linked to *only* this TS
func TargetServiceToggleActiveHandler(w http.ResponseWriter, r *http.Request) {

	// package name is last part of the URL
	tsname := r.URL.Path[len("/package/toggleactive/"):]

	// retrieve Target Service Struct
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// retrieve records with thats TS
	records, err := models.EbooksGetByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// change "active" bool in those records
	// and save each to DB
	for _, v := range records {
		if myTS.TSActive {
			v.Active = false
		} else {
			v.Active = true
		}
		_, vUpdateErr := models.EbookUpdate(v)
		if vUpdateErr != nil {
			logger.Error.Printf("can't update record %v: %v", v.ID, vUpdateErr)
		}
	}

	// change "active" bool in TS struct
	if myTS.TSActive {
		myTS.TSActive = false
	} else {
		myTS.TSActive = true
	}

	// save TS to DB
	tsUpdateErr := models.TSUpdate(myTS)
	if tsUpdateErr != nil {
		logger.Error.Println(tsUpdateErr)
	}

	// refresh TS page
	urlStr := "/package/" + tsname
	http.Redirect(w, r, urlStr, 303)

}
