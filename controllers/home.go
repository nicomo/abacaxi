package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/views"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})
	views.RenderTmpl(w, "home", d)
}
