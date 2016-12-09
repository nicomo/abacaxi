package controllers

import (
	"net/http"

	"github.com/nicomo/EResourcesMetadataHub/logger"
	"github.com/nicomo/EResourcesMetadataHub/models"
	"github.com/nicomo/EResourcesMetadataHub/views"
)

func TargetServiceHandler(w http.ResponseWriter, r *http.Request) {
	// our messages (errors, confirmation, etc) to the user & the template will be store in this map
	d := make(map[string]interface{})

	// package name is last part of the URL
	tsname := r.URL.Path[len("/package/"):]
	d["myPackage"] = tsname

	count := models.TSCountEbooks(tsname)
	d["myPackageEbooksCount"] = count

	if count > 0 { // no need to query for actual ebooks otherwise

		// how many ebooks have marc records
		nbMarcRecords := models.TSCountMarcRecords(tsname)
		logger.Debug.Println(nbMarcRecords)
		d["myPackageMarcRecordsCount"] = nbMarcRecords

		// how many ebooks have a PPN from the Sudoc Union Catalog
		nbPPNs := models.TSCountPPNs(tsname)
		logger.Debug.Println(nbPPNs)
		d["myPackagePPNsCount"] = nbPPNs

		// get the ebooks
		records, err := models.EbooksGetByPackageName(tsname)
		if err != nil {
			logger.Error.Println(err)
		}
		logger.Debug.Println("TargetServiceHandler", records)
		d["myRecords"] = records
	}

	views.RenderTmpl(w, "targetservice", d)
}
