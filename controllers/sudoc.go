package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/sudoc"
)

// SudocI2PHandler manages the consuming of a web service to retrieve a Sudoc ID
//  There's a "priority" isbn, we try to get a marc record number for this one first
// using the other isbns only if we can't
func SudocI2PHandler(w http.ResponseWriter, r *http.Request) {

	// data to be display in UI will be stored in this map
	// d := make(map[string]interface{})

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/sudoci2p/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	// generate the web service url for this record
	i2purl := sudoc.GenI2PURL(myEbook)

	// get PPN for i2purl
	result := sudoc.FetchPPN(i2purl)
	if result.Err != nil {
		logger.Error.Println(result.Err)
	}

	myEbook.Ppns = result.PPNs

	// actually save updated ebook struct to DB
	var ebkUpdateErr error
	myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
	if ebkUpdateErr != nil {
		logger.Error.Println(ebkUpdateErr)
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/ebook/" + ebookID
	http.Redirect(w, r, urlStr, 303)
}

// SudocI2PTSNewHandler retrieves PPNs for all ebooks linked to a Target Service that don't currently have one
func SudocI2PTSNewHandler(w http.ResponseWriter, r *http.Request) {

	// Target Service name is last part of the URL
	tsname := r.URL.Path[len("/sudoci2p-ts-new/"):]
	//d["myPackage"] = tsname

	records, err := models.EbooksGetNoPPNByTSName(tsname)
	if err != nil {
		logger.Error.Println(err)
	}

	// set up the pipeline
	in := sudoc.GenChannel(records)

	// fan out to 2 workers
	c1 := sudoc.CrawlPPN(in)
	c2 := sudoc.CrawlPPN(in)

	// fan in results
	for n := range sudoc.MergePPN(c1, c2) {
		logger.Debug.Println(n)
	}

}

// GetRecordHandler manages http request to use sudoc web service to retrieve marc records
func GetRecordHandler(w http.ResponseWriter, r *http.Request) {

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/sudocgetrecord/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	// ranging over the PPNs in a given local record
	// we fetch the sudoc marc record, and stop as soon as we get one
	for _, v := range myEbook.Ppns {
		record, err := sudoc.GetRecord("http://www.sudoc.fr/" + v + ".abes")
		if err != nil {
			logger.Error.Println(err)
			continue
		}

		if record != "" {

			// if the local record already has a mark record, update using delete / insert on the struct
			myEbook.MarcRecords = nil
			myEbook.MarcRecords = append(myEbook.MarcRecords, record)

			// actually save updated ebook struct to DB
			var ebkUpdateErr error
			myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
			if ebkUpdateErr != nil {
				logger.Error.Println(ebkUpdateErr)
			}

			if len(myEbook.MarcRecords) > 0 {
				break
			}
		}
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/ebook/" + ebookID
	http.Redirect(w, r, urlStr, 303)
}
