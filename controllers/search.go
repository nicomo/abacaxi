package controllers

import (
	"net/http"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// SearchHandler manages http requests throught the nav bar search form
func SearchHandler(w http.ResponseWriter, r *http.Request) {

	// results & messages to display in UI to be stored in this map
	d := make(map[string]interface{})

	result, searchterms, err := models.Search(r)
	if err != nil {
		logger.Error.Println(err)
	}
	d["myRecords"] = result
	d["searchterms"] = searchterms
	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "searchresults", d)

}
