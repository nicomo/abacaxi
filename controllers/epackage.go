package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
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

	if count > 0 { // no need to query for actual ebooks otherwise
		nbMarcRecords := models.PackageCountMarcRecords(packname)
		logger.Debug.Println(nbMarcRecords)
		d["myPackageMarcRecordsCount"] = nbMarcRecords
	}

	views.RenderTmpl(w, "epackage", d)
}
