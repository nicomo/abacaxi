package controllers

import (
	"net/http"

	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

type userMessages map[string]interface{}

// HomeHandler manages http requests on the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// various stats about the data in the DB
	ebksCount := models.EbooksCount()
	d["ebksCount"] = ebksCount

	ppnCount := models.EbooksCountPPNs()
	d["ppnCount"] = ppnCount

	unimarcCount := models.EbooksCountUnimarc()
	d["unimarcCount"] = unimarcCount

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	d["TSCount"] = len(TSListing)

	views.RenderTmpl(w, "home", d)
}
