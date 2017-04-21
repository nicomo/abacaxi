package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/sudoc"
	"github.com/nicomo/abacaxi/views"
)

// SudocI2PHandler manages the consuming of a web service to retrieve a Sudoc ID
func SudocI2PHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	// generate the web service url for this record
	i2purl, err := sudoc.GenI2PURL(myRecord.Identifiers)
	if err != nil {
		logger.Error.Println(err)
	}

	// get PPN for i2purl
	result := sudoc.FetchPPN(i2purl)
	if result.Err != nil {
		logger.Error.Println(result.Err)
	}

	// update live record with PPNs
	for _, v := range result.PPNs {
		var exists bool
		for _, w := range myRecord.Identifiers {
			if v == w.Identifier {
				exists = true
				continue
			}
		}
		if !exists {
			newPPN := models.Identifier{Identifier: v, IDType: models.IDTypePPN}
			myRecord.Identifiers = append(myRecord.Identifiers, newPPN)
		}
	}

	// actually save updated record struct to DB
	var ErrRecordUpdate error
	myRecord, ErrRecordUpdate = models.RecordUpdate(myRecord)
	if ErrRecordUpdate != nil {
		logger.Error.Println(ErrRecordUpdate)
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, 303)
}

// SudocI2PTSHandler retrieves PPNs for all records linked to a Target Service that don't currently have one
func SudocI2PTSHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})
	d["i2pType"] = "Get Sudoc PPN Record IDs for records currently without one"

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myPackage"] = tsname

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

// GetRecordHandler manages http request to use sudoc web service to retrieve marc record for 1 given ebook
func GetRecordHandler(w http.ResponseWriter, r *http.Request) {

	// retrieve the record ID from the request
	vars := mux.Vars(r)
	recordID := vars["recordID"]

	myRecord, err := models.RecordGetByID(recordID)
	if err != nil {
		logger.Error.Println(err)
	}

	// ranging over the PPNs in a given local record
	// we fetch the sudoc marc record, and stop as soon as we get one
	for _, v := range myRecord.Identifiers {
		if v.IDType != models.IDTypePPN {
			continue
		}
		record, err := sudoc.GetRecord("http://www.sudoc.fr/" + v.Identifier + ".abes")
		if err != nil {
			logger.Error.Println(err)
			continue
		}

		if record != "" {

			myRecord.RecordUnimarc = record

			// actually save updated ebook struct to DB
			var ErrRecordUpdate error
			myRecord, ErrRecordUpdate = models.RecordUpdate(myRecord)
			if ErrRecordUpdate != nil {
				logger.Error.Println(ErrRecordUpdate)
			}

			if len(myRecord.RecordUnimarc) > 0 {
				break
			}
		}
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/record/" + recordID
	http.Redirect(w, r, urlStr, 303)
}

// GetRecordsTSHandler retrieves Unimarc Records from Sudoc for all local records using a given target service
func GetRecordsTSHandler(w http.ResponseWriter, r *http.Request) {

	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// retrieve the Target Service from the request
	vars := mux.Vars(r)
	tsname := vars["targetservice"]
	d["myPackage"] = tsname

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
