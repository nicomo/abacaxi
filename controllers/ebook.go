package controllers

import (
	"net/http"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/views"
)

// EbookHandler displays a single record
func EbookHandler(w http.ResponseWriter, r *http.Request) {
	// data to be display in UI will be stored in this map
	d := make(map[string]interface{})

	// record ID is last part of the URL
	ebookId := r.URL.Path[len("/ebook/"):]

	myEbook, err := models.EbookGetById(ebookId)
	if err != nil {
		logger.Error.Println(err)
	}

	// format the dates
	d["formattedDateCreated"] = myEbook.DateCreated.Format(time.RFC822)
	if myEbook.DateUpdated.IsZero() {
		d["formattedDateUpdated"] = myEbook.DateUpdated.Format(time.RFC822)
	}

	d["Ebook"] = myEbook

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "ebook", d)
}

// EbookDeleteHandler handles deleting a single ebook
func EbookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// record ID is last part of the URL
	ebookId := r.URL.Path[len("/ebook/delete/"):]

	err := models.EbookDelete(ebookId)
	if err != nil {
		logger.Error.Println(err)

		// TODO: transmit either error or success message to user

		// redirect to home
		redirectURL := "/ebook/" + ebookId
		http.Redirect(w, r, redirectURL, 303)
	}

	// redirect to home
	http.Redirect(w, r, "/", 303)
}
