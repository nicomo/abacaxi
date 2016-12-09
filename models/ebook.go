// Package models stores the structs for the objects we have
package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type Ebook struct {
	Id                   bson.ObjectId `bson:"_id,omitempty"`
	DateCreated          time.Time
	DateUpdated          time.Time `bson:",omitempty"`
	Active               bool
	SfxId                string          `bson:",omitempty"`
	SFXLastHarvest       time.Time       `bson:",omitempty"`
	PublisherLastHarvest time.Time       `bson:",omitempty"`
	SudocLastHarvest     time.Time       `bson:",omitempty"`
	Authors              []string        `bson:",omitempty"`
	Title                string          `bson:",omitempty"`
	Publisher            string          `bson:",omitempty"`
	Pubdate              string          `bson:",omitempty"`
	Edition              int             `bson:",omitempty"`
	Lang                 string          `bson:",omitempty"`
	TargetService        []TargetService `bson:",omitempty"` // this is the name of the package in SFX, e.g. CAIRN QSJ
	OpenURL              string          `bson:",omitempty"`
	PackageURL           string          `bson:",omitempty"`
	Acquired             bool            `bson:",omitempty"`
	Isbns                []Isbn          `bson:",omitempty"`
	Ppns                 []PPN           `bson:",omitempty"`
	MarcRecords          []string        `bson:",omitempty"`
	Deleted              bool
}

type Isbn struct {
	Isbn       string
	Electronic bool
	Primary    bool
}

type PPN struct {
	Ppn        string
	Electronic bool
	Primary    bool
}

// EbookCreate saves a single ebook to mongo DB
func EbookCreate(ebk Ebook) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getEbooksColl()

	// let's add the time and save
	ebk.DateCreated = time.Now()
	err := coll.Insert(ebk)
	if err != nil {
		logger.Error.Printf("could not save ebook with isbn %v in DB: %s", ebk.Isbns[0], err)
		return err
	}

	return nil
}

// EbookGetByIsbn retrieves an ebook
func EbookGetByIsbns(isbns []string) (Ebook, error) {
	ebk := Ebook{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	// construct query

	var isbnQry []bson.M
	for i := 0; i < len(isbns); i++ {
		if isbns[i] != "" {
			sTerm := bson.M{"isbns.isbn": isbns[i]}
			isbnQry = append(isbnQry, sTerm)
		}
	}
	qry := bson.M{
		"$or": isbnQry,
	}

	// execute query
	err := coll.Find(qry).One(&ebk)
	if err != nil {
		return ebk, err
	}

	return ebk, nil
}

// EbooksCreateOrUpdate checks if ebook exists in DB, using ISBN, then routes to either create or update
func EbooksCreateOrUpdate(records []Ebook) (int, int, error) {

	var createdCounter, updatedCounter int

	for _, record := range records { // for each record

		// pull out the isbns
		isbnsToQuery := make([]string, 0)
		for _, isbn := range record.Isbns { // for each isbn
			if isbn.Isbn != "" {
				isbnsToQuery = append(isbnsToQuery, isbn.Isbn)
			}
		}

		// test if we already know this ebook
		existingRecord, err := EbookGetByIsbns(isbnsToQuery)
		if err != nil { // we don't: none of the isbns were found in DB
			// let's create a new record
			ebkCreateErr := EbookCreate(record)
			if ebkCreateErr != nil {
				logger.Error.Println(ebkCreateErr)
			}
			createdCounter++
			continue
		}

		// TODO: we've found the record, let's update.it
		fmt.Println(existingRecord)
		updatedCounter++
	}
	return createdCounter, updatedCounter, nil
}

//TODO: EbookExists returns bool. See https://godoc.org/gopkg.in/mgo.v2#Query.Count

//TODO: EbookUpdate
func EbookUpdate(ebk Ebook) (Ebook, error) {
	return ebk, nil
}

//TODO: EbookSoftDelete
func EbookSoftDelete(ebkId int) error {
	return nil
}

//TODO: EbookDelete
func EbookDelete(ebkId int) error {
	return nil
}

//TODO : EbooksGetByPackageName
func EbooksGetByPackageName(tsname string) ([]Ebook, error) {
	result := []Ebook{}
	return result, nil
}

//TODO: EbooksGetByTitle
