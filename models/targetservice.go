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
	ID                     bson.ObjectId `bson:"_id,omitempty"`
	TSName                 string        `bson:",omitempty" schema:"tsname"`
	TSDisplayName          string        `bson:",omitempty" schema:"tsdisplayname"`
	TSDateCreated          time.Time
	TSDateUpdated          time.Time `bson:",omitempty"`
	TSPublisherLastHarvest time.Time `bson:",omitempty"`
	TSSFXLastHarvest       time.Time `bson:",omitempty"`
	TSSudocLastHarvest     time.Time `bson:",omitempty"`
	TSActive               bool      `schema:"tsactive"`
	TSCsvConf              TSCSVConf `bson:",omitempty"`
}

// TSCSVConf indicates the # of fields + column (index) of the various pieces of info in the csv file
type TSCSVConf struct {
	Col1  string `bson:",omitempty" tag_col:"0"`
	Col2  string `bson:",omitempty" tag_col:"1"`
	Col3  string `bson:",omitempty" tag_col:"2"`
	Col4  string `bson:",omitempty" tag_col:"3"`
	Col5  string `bson:",omitempty" tag_col:"4"`
	Col6  string `bson:",omitempty" tag_col:"5"`
	Col7  string `bson:",omitempty" tag_col:"6"`
	Col8  string `bson:",omitempty" tag_col:"7"`
	Col9  string `bson:",omitempty" tag_col:"8"`
	Col10 string `bson:",omitempty" tag_col:"9"`
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

// TSCountRecordsUnimarc counts how many records for this package have proper MARC Records
func TSCountRecordsUnimarc(tsname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice.tsname": tsname, "recordunimarc": bson.M{"$ne": nil}})
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
//NOTE: should review the code generally to see when to really use pointers rather than values
// here : should pbly be a value, since we neither change nor return the struct
func TSCreate(ts TargetService) error {

	// TODO: date created not properly saved
	ts.TSDateCreated = time.Now()

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	err := coll.Insert(&ts)
	if err != nil {
		logger.Error.Println(err)
		if mgo.IsDup(err) { // this Target service already exists in DB
			ErrTSIsDup := errors.New("Target service " + ts.TSName + " already exists")
			return ErrTSIsDup
		}
		return err
	}
	logger.Info.Printf("Created a new Target Service: %s", ts.TSName)
	return nil
}

// TSUpdate updates a target service
func TSUpdate(ts TargetService) error {

	ts.TSDateUpdated = time.Now()

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
