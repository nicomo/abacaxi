package controllers

import (
	"net/http"

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

	d["Ebook"] = myEbook

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "ebook", d)
}

//TODO: EbookDeleteHandler handles deleting a single ebook
func EbookDeleteHandler(w http.ResponseWriter, r *http.Request) {}

// TODO: EbookUpdateHandler handles updating a single ebook
func EbookUpdateHandler(w http.ResponseWriter, r *http.Request) {}

//TODO: EbooksHandler shows a list of ebooks
func EbooksHandler(w http.ResponseWriter, r *http.Request) {}
