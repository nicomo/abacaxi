package controllers

import (
	"net/http"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

// ReportsHandler retrieves and displays the last 100 reports for batch operations
func ReportsHandler(w http.ResponseWriter, r *http.Request) {
	d := make(map[string]interface{})

	reports, err := models.ReportsGet()
	if err != nil {
		logger.Error.Println(err)
	}

	d["reports"] = reports

	// list of existing TargetServices to be displayed in nav.
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "reports", d)
}
