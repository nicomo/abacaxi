// Package models stores the structs for the objects we have
package models

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/abacaxi/logger"
)

// TargetService represents an SFX Target Service,
// i.e. a package with its provider
// e.g. SPRINGER MATH EBOOKS
type TargetService struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Name        string        `bson:",omitempty" schema:"name"`
	DisplayName string        `bson:",omitempty" schema:"displayname"`
	DateCreated time.Time
	DateUpdated time.Time `bson:",omitempty"`
	Active      bool      `schema:"active"`
}

// GetTargetService retrieves a target service
func GetTargetService(tsname string) (TargetService, error) {

	ts := TargetService{}
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	qry := bson.M{"name": tsname}
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

// TSCountRecords counts the number of records for this target service
func TSCountRecords(tsname string) int {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection records
	coll := getRecordsColl()

	//  query records by target service name, aka Target Service in SFX (and in models.Records struct)
	qry := coll.Find(bson.M{"targetservices.name": tsname})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// TSCountRecordsUnimarc counts how many records for this target service have proper MARC Records
func TSCountRecordsUnimarc(tsname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection records
	coll := getRecordsColl()

	//  query records by target service name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservices.name": tsname, "recordunimarc": bson.M{"$ne": nil}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// TSCountPPNs counts how many records for this target service have proper PicaPublication Numbers coming from ABES
func TSCountPPNs(tsname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks by target service name, aka Target Service in SFX (and in models.Ebook struct) and checks if PPN exists
	qry := coll.Find(bson.M{"targetservices.name": tsname, "identifiers.idtype": IDTypePPN})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

//TSCreate registers a new target service, aka ebook package in mongo db
//NOTE: should review the code generally to see when to really use pointers rather than values
// here : should pbly be a value, since we neither change nor return the struct
func TSCreate(ts TargetService) error {

	// TODO: date created not properly saved
	ts.DateCreated = time.Now()

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	err := coll.Insert(&ts)
	if err != nil {
		logger.Error.Println(err)
		if mgo.IsDup(err) { // this Target service already exists in DB
			ErrTSIsDup := errors.New("Target service " + ts.Name + " already exists")
			return ErrTSIsDup
		}
		return err
	}
	logger.Info.Printf("Created a new Target Service: %s", ts.Name)
	return nil
}

// TSDelete removes a target service
func TSDelete(tsname string) error {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection records
	coll := getTargetServiceColl()

	// delete record
	qry := bson.M{"name": tsname}
	err := coll.Remove(qry)
	if err != nil {
		return err
	}

	return nil
}

// TSUpdate updates a target service
func TSUpdate(ts TargetService) error {

	ts.DateUpdated = time.Now()

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
