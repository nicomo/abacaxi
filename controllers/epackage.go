package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/views"
)

func EpackageHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// package name is last part of the URL
	packname := r.URL.Path[len("/package/"):]
	d["myPackage"] = packname

	count := models.PackageCountEbooks(packname)
	d["myPackageEbooksCount"] = count

	views.RenderTmpl(w, "epackage", d)
}
