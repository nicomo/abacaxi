package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/sudoc"
)

// SudocIsbn2PpnHandler manages the consuming of a web service to retrieve a Sudoc ID
//  There's a "priority" isbn, we try to get a marc record number for this one first
// using the other isbns only if we can't
func SudocIsbn2PpnHandler(w http.ResponseWriter, r *http.Request) {

	// data to be display in UI will be stored in this map
	d := make(map[string]interface{})

	// record ID is last part of the URL
	ebookId := r.URL.Path[len("/sudocisbn2ppn/"):]

	myEbook, err := models.EbookGetById(ebookId)
	if err != nil {
		logger.Error.Println(err)
	}

	// 2 URLs, 1 for the preferred isbn, 1 for the other isbns
	var priorityURL string
	var allIsbns []string
	allIsbnsURL := "http://www.sudoc.fr/services/isbn2ppn/"

	logger.Debug.Println(myEbook.Isbns)

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

	logger.Debug.Println("priorityURL..." + priorityURL)
	logger.Debug.Println("allIsbnsURL..." + allIsbnsURL)

	// retrieve PPN from sudoc web service :
	// preferred isbns 1st if there's one
	// other isbns if preferred gets no result (error received)
	// FIXME: logic is wrong, too much repetitions between priority isbn and other isbns
	if priorityURL != "" {
		priorityPPN, sudocErr := sudoc.FetchPPN(priorityURL)
		if sudocErr != nil {
			logger.Error.Println(sudocErr)
			d["sudocErr"] = sudocErr
		}
		myPPN := models.PPNCreate(priorityPPN, true)
		myEbook.Ppns = myPPN
		logger.Debug.Printf("%v", myEbook)

		// actually save updated ebook struct to DB
		var ebkUpdateErr error
		myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
		if ebkUpdateErr != nil {
			logger.Error.Println(ebkUpdateErr)
		}

	} else {
		allPPN, allSudocErr := sudoc.FetchPPN(allIsbnsURL)
		if allSudocErr != nil {
			logger.Error.Println(allSudocErr)
			d["allSudocErr"] = allSudocErr
		}
		myPPNs := models.PPNCreate(allPPN, false)
		myEbook.Ppns = myPPNs
		logger.Debug.Printf("%v", myEbook)

		// actually save updated ebook struct to DB
		var ebkUpdateErr error
		myEbook, ebkUpdateErr = models.EbookUpdate(myEbook)
		if ebkUpdateErr != nil {
			logger.Error.Println(ebkUpdateErr)
		}
	}

	// redirect to book detail page
	// TODO: transmit either error or success message to user
	urlStr := "/ebook/" + ebookId
	http.Redirect(w, r, urlStr, 303)

}
