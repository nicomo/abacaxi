// Package models stores the structs for the objects we have
package models

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/abacaxi/logger"
)

// TargetService represents an SFX Target Service,
// i.e. a package with its provider
// e.g. SPRINGER MATH EBOOKS
type TargetService struct {
	ID                     bson.ObjectId `bson:"_id,omitempty"`
	TSName                 string        `bson:",omitempty" schema:"tsname"`
	TSDisplayName          string        `bson:",omitempty" schema:"tsdisplayname"`
	TSDateCreated          time.Time
	TSDateUpdated          time.Time `bson:",omitempty"`
	TSPublisherLastHarvest time.Time `bson:",omitempty"`
	TSSFXLastHarvest       time.Time `bson:",omitempty"`
	TSSudocLastHarvest     time.Time `bson:",omitempty"`
	TSActive               bool      `schema:"tsactive"`
}

// GetTargetService retrieves a target service
func GetTargetService(tsname string) (TargetService, error) {

	ts := TargetService{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	qry := bson.M{"tsname": tsname}
	err := coll.Find(qry).One(&ts)
	if err != nil {
		return ts, err
	}
	return ts, nil

}

// GetTargetServicesListing retrieves the full list of target services
func GetTargetServicesListing() ([]TargetService, error) {

	var TSListing []TargetService

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	err := coll.Find(bson.M{}).Sort("tsname").All(&TSListing)
	if err != nil {
		return TSListing, err
	}

	return TSListing, nil

}

// TSCountEbooks counts the number of ebooks for this package
func TSCountEbooks(tsname string) int {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice.tsname": tsname})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// TSCountMarcRecords counts how many records for this package have proper MARC Records
func TSCountMarcRecords(tsname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice.tsname": tsname, "marcrecords": bson.M{"$ne": nil}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// TSCountPPNs counts how many records for this package have proper PicaPublication Numbers coming from ABES
func TSCountPPNs(tsname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct) and checks if PPN exists
	qry := coll.Find(bson.M{"targetservice.tsname": tsname, "ppns": bson.M{"$exists": true}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

//TSCreate registers a new target service, aka ebook package in mongo db
func TSCreate(r *http.Request) error {

	// init our Target Service struct
	ts := new(TargetService)

	// used by gorilla schema to parse html forms
	decoder := schema.NewDecoder()

	// we parse the form
	parseErr := r.ParseForm()
	logger.Info.Println(r.Form)
	if parseErr != nil {
		logger.Error.Println(parseErr)
		return parseErr
	}

	// r.PostForm is a map of our POST form values
	// we create a struct from form
	// but ignore the fields which do not exist in the struct
	decoder.IgnoreUnknownKeys(true)
	errDecode := decoder.Decode(ts, r.PostForm)
	if errDecode != nil {
		return errDecode
	}

	ts.TSDateCreated = time.Now()

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	err := coll.Insert(&ts)
	if err != nil {
		logger.Error.Println(err)
		if mgo.IsDup(err) { // this Target service already exists in DB
			tsIsDupErr := errors.New("Target service " + ts.TSName + " already exists")
			return tsIsDupErr
		}
		return err
	}

	return nil
}

// TSUpdate updates a target service
func TSUpdate(ts TargetService) error {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	// execute query
	err := coll.Update(bson.M{"_id": ts.ID}, ts)
	if err != nil {
		return err
	}

	return nil
}
