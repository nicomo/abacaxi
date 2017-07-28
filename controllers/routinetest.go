package controllers

import (
	"fmt"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/views"
)

type CountFeedback struct {
	CountDone    int
	CountDiscard int
	CountCreate  int
	CountUpdate  int
}

// RoutineTestHandler handles blabla
func RoutineTestHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in RoutineTestHandler")

	d := make(map[string]interface{})
	d["working"] = "working..."
	reports, err := models.ReportsGet()
	if err != nil {
		logger.Error.Println(err)
	}
	d["reports"] = reports
	views.RenderTmpl(w, "wsform2", d)

}

func FormGetHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in FormGetHandler")
	views.RenderTmpl(w, "wsform", nil)
}

func FormPostHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in FormPostHandler")

	r.ParseForm()

	// create the counter
	myCounters := CountFeedback{
		CountDone:    0,
		CountDiscard: 0,
		CountCreate:  0,
		CountUpdate:  0,
	}

	go myCounters.testReport()

	// do something with the values received from the form
	for k, v := range r.Form {
		fmt.Printf("k: %v / v: %v\n", k, v)
	}

	http.Redirect(w, r, "routinetest", http.StatusFound)

}

func (myCF *CountFeedback) testReport() {
	logger.Debug.Println("in testReport")
	for {
		time.Sleep(1 * time.Second)
		if myCF.CountDone < 10 {
			logger.Debug.Println(myCF.CountDone)
			myCF.CountDone += 2
		} else {
			break
		}
	}

	report := models.Report{
		ID:          bson.NewObjectId(),
		DateCreated: time.Now(),
		ReportType:  models.UploadCsv,
		Text:        "some lame text for now",
	}

	if err := report.ReportCreate(); err != nil {
		logger.Error.Println(err)
	}

}
