// Package models stores the structs for the objects we have
package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

type Ebook struct {
	DateCreated          time.Time
	DateUpdated          time.Time
	Active               bool
	SfxId                string
	SFXLastHarvest       time.Time
	PublisherLastHarvest time.Time
	SudocLastHarvest     time.Time
	Authors              []string
	Title                string
	Publisher            string
	Pubdate              string
	Edition              int
	Lang                 string
	TargetService        string
	OpenURL              string
	PackageURL           string
	Acquired             bool
	Isbns                []Isbn
	Ppns                 []PPN
	MarcRecord           []string
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

//TODO: EbookCreate
func EbookCreate(ebk Ebook) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getEbooksCol()

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
func EbookGetByIsbn(isbn string) (Ebook, error) {
	ebk := Ebook{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()
	err := coll.Find(bson.M{"isbn": isbn}).One(&ebk)
	if err != nil {
		return ebk, err
	}

	return ebk, nil
}

//TODO: EbooksCreateOrUpdate
func EbooksCreateOrUpdate(records []Ebook) error {

	for _, record := range records { // for each record
		for _, isbn := range record.Isbns { // for each isbn
			if isbn.Isbn != "" {
				workingRecord, err := EbookGetByIsbn(isbn.Isbn) // test if we already know this ebook
				if err != nil {                                 // we don't: isbn not found
					// let's create a new record
					ebkCreateErr := EbookCreate(record)
					if ebkCreateErr != nil {
						logger.Error.Println(ebkCreateErr)
						return ebkCreateErr
					}
				}

				// TODO: we've found the record, let's update.it
				fmt.Println("workingRecord", workingRecord)
			}
		}
	}
	return nil
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
