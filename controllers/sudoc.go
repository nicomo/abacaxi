package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/views"

	"github.com/nicomo/EResourcesMetadataHub/sudoc"
)

// GetPPNHandler manages the consuming of a web service to retrieve a Sudoc ID
//  There's a "priority" isbn, we try to get a marc record number for this one first
// using the other isbns only if we can't
func GetPPNHandler(w http.ResponseWriter, r *http.Request) {

	// data to be display in UI will be stored in this map
	d := make(map[string]interface{})

	// record ID is last part of the URL
	ebookId := r.URL.Path[len("/getppn/"):]

	myEbook, err := models.EbookGetById(ebookId)
	if err != nil {
		logger.Error.Println(err)
	}

	var priorityURL string
	var allIsbns []string
	allIsbnsURL := "http://www.sudoc.fr/services/isbn2ppn/"

	logger.Debug.Println(myEbook.Isbns)

	for _, v := range myEbook.Isbns {
		if v.Primary {
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

	if priorityURL != "" {
		_, sudocErr := sudoc.FetchPPN(priorityURL)
		if sudocErr != nil {
			logger.Error.Println(sudocErr)
			_, allSudocErr := sudoc.FetchPPN(allIsbnsURL)
			if allSudocErr != nil {
				logger.Error.Println(allSudocErr)
				d["sudocErr"] = allSudocErr
			}
		}
	}

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "ebook", d)

}
