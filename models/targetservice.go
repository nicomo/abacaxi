// Package models stores the structs for the objects we have
package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type TargetService struct {
	Id                     bson.ObjectId `bson:"_id,omitempty"`
	TSName                 string
	TSPublisherLastHarvest time.Time `bson:",omitempty"`
	TSSFXLastHarvest       time.Time `bson:",omitempty"`
	TSSudocLastHarvest     time.Time `bson:",omitempty"`
	TSActive               bool
}

// getTargetService retrieves a target service
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

// TODO: updateTargetService

// TSCountEbooks counts the number of ebooks for this package
func TSCountEbooks(tsname string) int {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": tsname})
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
	qry := coll.Find(bson.M{"targetservice": tsname, "marcrecords": bson.M{"$ne": nil}})
	logger.Debug.Println(qry)
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

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": tsname, "ppns": false})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// TSUpdate updates a target service
func TSUpdate(ts TargetService) error {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getTargetServiceColl()

	// execute query
	err := coll.Update(bson.M{"_id": ts.Id}, ts)
	if err != nil {
		return err
	}

	return nil
}
