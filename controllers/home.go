package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/views"
)

type userMessages map[string]interface{}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTmpl(w, "home", nil)
}
