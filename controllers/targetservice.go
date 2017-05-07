package controllers

import (
	"context"
	"net/http"
	"strconv"

	"log"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

// getPrevNext takes the number of records in a result
// returns the records to skip to when we split the result set in chunks of n
func getPrevNext(page int, count int) (int, int) {

	// small batch, no need to paginate
	if count < 100 {
		return 0, 0
	}

	// at the end : next = 0, previous = page -100
	if count-page < 100 {
		return page - 200, 0
	}

	// at the beginning
	if page-100 < 0 {
		return -1, 200
	}
	return page - 100, page + 100
}

func getTSNameAndPage(r *http.Request) (string, int) {

	// check if we have a tsname + page coming in the Request context
	if ctxTSName, page, ok := fromContextPage(r.Context()); ok {
		return ctxTSName, page
	}

	// the name of the target service we're interested in is in the router variables
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	return tsname, 0
}

// createTSStructFromForm creates a TS struct from a form
func createTSStructFromForm(r *http.Request) (models.TargetService, error) {
	// init our Target Service struct
	ts := models.TargetService{}

	// used by gorilla schema to parse html forms
	decoder := schema.NewDecoder()

	// we parse the form
	ErrParse := r.ParseForm()
	if ErrParse != nil {
		logger.Error.Println(ErrParse)
		return ts, ErrParse
	}
	// r.PostForm is a map of our POST form values
	// we create a struct from form
	// but ignore the fields which would not exist in the struct
	decoder.IgnoreUnknownKeys(true)
	errDecode := decoder.Decode(&ts, r.PostForm)
	if errDecode != nil {
		logger.Error.Println(errDecode)
		return ts, errDecode
	}

	return ts, nil
}

// TargetServiceHandler retrieves the ebooks linked to a Target Service
//  and various other info, e.g. number of library records linked, etc.
func TargetServiceHandler(w http.ResponseWriter, r *http.Request) {

	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// Get session
	sess := session.Instance(r)
	if sess.Values["id"] != nil {
		d["IsLoggedIn"] = true
	}

	tsname, page := getTSNameAndPage(r)
	d["myPackage"] = tsname

	// get the TS Struct from DB
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}
	d["IsTSActive"] = myTS.TSActive

	// convert the csv configuration into a string to be displayed
	d["myPackageCSVConf"] = csvConf2String(myTS.TSCsvConf)

	// any local records records have this TS?
	count := models.TSCountRecords(tsname)
	d["myPackageRecordsCount"] = count

	if count > 0 { // no need to query for actual local records otherwise

		// if we need to paginate, get record skip integers, e.g. skip to records 20, 40, 60, etc;
		// to be used by mgo.skip() to do a simple paginate
		previous, next := getPrevNext(page, count)
		if previous >= 0 {
			d["previous"] = previous
		}
		if next != 0 {
			d["next"] = next
		}

		// how many local records have marc records
		nbRecordsUnimarc := models.TSCountRecordsUnimarc(tsname)
		d["myPackageRecordsUnimarcCount"] = nbRecordsUnimarc

		// how many local records have a PPN from the Sudoc Union Catalog
		nbPPNs := models.TSCountPPNs(tsname)
		d["myPackagePPNsCount"] = nbPPNs

		// get the records
		records, err := models.RecordsGetByTSName(tsname, page)
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

// TargetServiceExportKbartHandler exports a batch of records as a KBART-compliant .csv file
func TargetServiceExportKbartHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve tsname passed in url
	vars := mux.Vars(r)
	tsname := vars["targetservice"]

	// get the relevant records
	records, err := models.RecordsGetWithUnimarcByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
		//TODO: exit cleanly with user message on error
		log.Fatalln(err)
	}

	filename := tsname + ".csv"

	// create .csv kbart file
	filesize, err := models.CreateKbartFile(records, filename)
	if err != nil {
		logger.Error.Panicf("could not create Kbart file: %v", err)
		//TODO: exit cleanly with user message on error
		log.Fatalln(err)
	}

	// exporting the created file
	if err := exportFile(w, r, filename, filesize); err != nil {
		logger.Error.Printf("couldn't stream the export file: %v", err)
	}

}

// TargetServiceExportUnimarcHandler exports a batch of unimarc records
func TargetServiceExportUnimarcHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve TS name
	vars := mux.Vars(r)
	tsname := vars["targetservice"]

	// get the relevant records
	records, err := models.RecordsGetWithUnimarcByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
		//TODO: exit cleanly with user message on error
		panic(err)
	}

	filename := tsname + ".xml"

	// create the file
	filesize, err := models.CreateUnimarcFile(records, filename)
	if err != nil {
		logger.Error.Printf("could not create file: %v", err)
		//TODO: exit cleanly with user message on error
	}

	// export the file
	if err := exportFile(w, r, filename, filesize); err != nil {
		logger.Error.Printf("couldn't stream the export file: %v", err)
	}

}

// TargetServiceUpdateGetHandler fills the update form for a Target Service
func TargetServiceUpdateGetHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// the name of the target service we're interested in is in the router variables
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myPackage"] = tsname

	// retrieve Target Service Struct
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	d["myTS"] = myTS

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "tsupdate", d)
}

// TargetServicePageHandler gets TS with list of books for page n
func TargetServicePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		logger.Error.Println(err)
	}
	tsname := vars["targetservice"]

	// insert the page in the http.Request Context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	ctx = newContextPage(ctx, page)
	ctx = newContextTSName(ctx, tsname)

	// redirect to upload get page
	TargetServiceHandler(w, r.WithContext(ctx))

}

// TargetServiceUpdatePostHandler updates a target service
func TargetServiceUpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	d := make(map[string]interface{})

	// the name of the target service we're interested in is in the router variables
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myPackage"] = tsname

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	ts, ErrForm := createTSStructFromForm(r)
	if ErrForm != nil {
		d["ErrTSUpdate"] = ErrForm
		logger.Error.Println(ErrForm)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	if ts.TSDisplayName == "" {
		d["ErrTSUpdate"] = "Display name can't be empty for TS " + tsname
		logger.Info.Println("Display name can't be empty for TS " + tsname)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	if !csvConfValidate(ts.TSCsvConf) {
		csvConfNotValid := "csv configuration not valid for TS " + tsname + " : should have a title and isbn/e-isbn"
		d["ErrTSUpdate"] = csvConfNotValid
		logger.Info.Println(csvConfNotValid)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	tsToUpdate, ErrTsToUpdate := models.GetTargetService(tsname)
	if ErrTsToUpdate != nil {
		logger.Error.Println(ErrTsToUpdate)
		d["ErrTSUpdate"] = ErrTsToUpdate
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	ts.ID = tsToUpdate.ID

	err := models.TSUpdate(ts)
	if err != nil {
		d["ErrTSUpdate"] = err
		logger.Error.Println(err)
		views.RenderTmpl(w, "tsupdate", d)
		return
	}

	redirectURL := "/package/" + tsname
	http.Redirect(w, r, redirectURL, 303)
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

	ts, ErrForm := createTSStructFromForm(r)
	if ErrForm != nil {
		d["tsCreateErr"] = ErrForm
		logger.Error.Println(ErrForm)
		views.RenderTmpl(w, "targetservicenewget", d)
		return
	}

	if !csvConfValidate(ts.TSCsvConf) {
		csvConfNotValid := "csv configuration not valid : should have a title and isbn/e-isbn"
		d["tsCreateErr"] = csvConfNotValid
		views.RenderTmpl(w, "targetservicenewget", d)
		return
	}

	err := models.TSCreate(ts)
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

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]

	// retrieve Target Service Struct
	myTS, err := models.GetTargetService(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// retrieve records with thats TS
	records, err := models.RecordsGetByTSName(tsname, 0)
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
		_, ErrRecordUpdate := models.RecordUpdate(v)
		if ErrRecordUpdate != nil {
			logger.Error.Printf("can't update record %v: %v", v.ID, ErrRecordUpdate)
		}
	}

	// change "active" bool in TS struct
	if myTS.TSActive {
		myTS.TSActive = false
	} else {
		myTS.TSActive = true
	}

	// save TS to DB
	ErrTSUpdate := models.TSUpdate(myTS)
	if ErrTSUpdate != nil {
		logger.Error.Println(ErrTSUpdate)
	}

	// refresh TS page
	urlStr := "/package/" + tsname
	http.Redirect(w, r, urlStr, 303)
}
