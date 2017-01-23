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
	d := make(map[string]interface{})

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/sudoci2p/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	// 2 URLs, 1 for the preferred isbn, 1 for the other isbns
	var priorityURL string
	var allIsbns []string
	allIsbnsURL := "http://www.sudoc.fr/services/isbn2ppn/"

	for _, v := range myEbook.Isbns {
		if v.Electronic {
			priorityURL = "http://www.sudoc.fr/services/isbn2ppn/" + v.Isbn
		} else {
			allIsbns = append(allIsbns, v.Isbn)
		}
	}

	for i, v := range allIsbns {
		if i < len(allIsbns)-1 {
			allIsbnsURL = allIsbnsURL + v + ","
		}
		allIsbnsURL = allIsbnsURL + v
	}

	// retrieve PPN from sudoc web service :
	// preferred isbns 1st if there's one
	// other isbns if preferred gets no result (error received)
	// FIXME: logic is wrong, too much repetitions between priority isbn and other isbns
	if priorityURL != "" {
		result := sudoc.FetchPPN(priorityURL)
		if result.Err != nil {
			logger.Error.Println(result.Err)
			d["sudocErr"] = result.Err
		}
		myPPN := models.PPNCreate(result.PPNs, true)
		myEbook.Ppns = myPPN

		// actually save updated ebook struct to DB
		var ebkUpdateErr error
		myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
		if ebkUpdateErr != nil {
			logger.Error.Println(ebkUpdateErr)
		}
	}

	if len(myEbook.Ppns) == 0 {
		result := sudoc.FetchPPN(allIsbnsURL)
		if result.Err != nil {
			logger.Error.Println(result.Err)
			d["allSudocErr"] = result.Err
		}
		myPPNs := models.PPNCreate(result.PPNs, false)
		myEbook.Ppns = myPPNs

		// actually save updated ebook struct to DB
		var ebkUpdateErr error
		myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
		if ebkUpdateErr != nil {
			logger.Error.Println(ebkUpdateErr)
		}
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/ebook/" + ebookID
	http.Redirect(w, r, urlStr, 303)
}

// SudocI2PTSNewHandler retrieves PPNs for all ebooks linked to a Target Service that don't currently have one
func SudocI2PTSNewHandler(w http.ResponseWriter, r *http.Request) {

	// package name is last part of the URL
	//tsname := r.URL.Path[len("/sudoci2p-ts-new/"):]
	//d["myPackage"] = tsname

	testEbks := []string{
		"http://www.sudoc.fr/services/isbn2ppn/9782869783836",
		"http://www.sudoc.fr/services/isbn2ppn/9782844506931",
		"http://www.sudoc.fr/services/isbn2ppn/9782760522213",
		"http://www.sudoc.fr/services/isbn2ppn/9782806226129",
		"http://www.sudoc.fr/services/isbn2ppn/9782100555727",
		"http://www.sudoc.fr/services/isbn2ppn/fakeisbn1",
		"http://www.sudoc.fr/services/isbn2ppn/fakeisbn2",
		"http://www.sudoc.fr/services/isbn2ppn/fakeisbn3",
		"http://www.sudoc.fr/services/isbn2ppn/fakeisbn4",
		"http://www.sudoc.fr/services/isbn2ppn/fakeisbn5",
	}

	// set up the pipeline
	in := sudoc.GenChannel(testEbks)

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

	// we put the "electronic" ppns 1st in the map, then the others
	// we sort the map
	// we fetch the record, and stop as soon as we get one
	var sortedPPNs []string
	for _, v := range myEbook.Ppns {
		if v.Electronic {
			sortedPPNs = append(sortedPPNs, "http://www.sudoc.fr/"+v.Ppn+".abes")
		}
	}
	for _, v := range myEbook.Ppns {
		if v.Electronic == false {
			sortedPPNs = append(sortedPPNs, "http://www.sudoc.fr/"+v.Ppn+".abes")
		}
	}

	for _, v := range sortedPPNs {
		record, err := sudoc.GetRecord(v)
		if err != nil {
			logger.Error.Println(err)
			continue
		}

		if record != "" {

			// if the local record already has a mark record, update using delete / insert
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
