package controllers

import (
	"net/http"

	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// HomeHandler manages http requests on the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing
	views.RenderTmpl(w, "home", d)
}
