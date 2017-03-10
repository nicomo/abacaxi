package controllers

import (
	"net/http"
	"time"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

// EbookHandler displays a single record
func EbookHandler(w http.ResponseWriter, r *http.Request) {
	// data to be display in UI will be stored in this map
	d := make(map[string]interface{})

	// Get session
	sess := session.Instance(r)
	if sess.Values["id"] != nil {
		d["IsLoggedIn"] = true
	}

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/ebook/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	// format the dates
	d["formattedDateCreated"] = myEbook.DateCreated.Format(time.RFC822)

	if !myEbook.DateUpdated.IsZero() {
		d["formattedDateUpdated"] = myEbook.DateUpdated.Format(time.RFC822)
	}

	d["Ebook"] = myEbook

	// list of TS appearing in menu
	TSListing, _ := models.GetTargetServicesListing()
	d["TSListing"] = TSListing

	views.RenderTmpl(w, "ebook", d)
}

// EbookDeleteHandler handles deleting a single ebook
func EbookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/ebook/delete/"):]

	err := models.EbookDelete(ebookID)
	if err != nil {
		logger.Error.Println(err)

		// TODO: transmit either error or success message to user

		// redirect to home
		redirectURL := "/ebook/" + ebookID
		http.Redirect(w, r, redirectURL, 303)
	}

	// redirect to home
	http.Redirect(w, r, "/", 303)
}

//EbookToggleAcquiredHandler toggles the boolean value "acquired" for an ebook
func EbookToggleAcquiredHandler(w http.ResponseWriter, r *http.Request) {

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/ebook/toggleacquired/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	if myEbook.Acquired {
		myEbook.Acquired = false
	} else {
		myEbook.Acquired = true
	}

	myEbook, err = models.EbookUpdate(myEbook)
	if err != nil {
		logger.Error.Println(err)
	}

	// refresh ebook page
	urlStr := "/ebook/" + ebookID
	http.Redirect(w, r, urlStr, 303)
}

// EbookToggleActiveHandler toggles the boolean value "active" for an ebook
func EbookToggleActiveHandler(w http.ResponseWriter, r *http.Request) {

	// record ID is last part of the URL
	ebookID := r.URL.Path[len("/ebook/toggleactive/"):]

	myEbook, err := models.EbookGetByID(ebookID)
	if err != nil {
		logger.Error.Println(err)
	}

	if myEbook.Active {
		myEbook.Active = false
	} else {
		myEbook.Active = true
	}

	myEbook, err = models.EbookUpdate(myEbook)
	if err != nil {
		logger.Error.Println(err)
	}

	// refresh ebook page
	urlStr := "/ebook/" + ebookID
	http.Redirect(w, r, urlStr, 303)
}
