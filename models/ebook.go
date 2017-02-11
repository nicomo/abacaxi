// Package models stores the structs for the objects we have & interacts with mongo
package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/abacaxi/logger"
)

// Ebook is what it says on the tin : struct for an ebook
type Ebook struct {
	ID                   bson.ObjectId `bson:"_id,omitempty"`
	DateCreated          time.Time
	DateUpdated          time.Time `bson:",omitempty"`
	Active               bool
	SFXID                string          `bson:",omitempty"`
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
	Ppns                 []string        `bson:",omitempty"`
	RecordUnimarc        string          `bson:",omitempty"`
	RecordMarc21         string          `bson:",omitempty"`
	Deleted              bool
}

// Isbn embedded in an ebook
type Isbn struct {
	Isbn       string
	Electronic bool
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

// EbookDelete deletes a single ebook from DB
func EbookDelete(ID string) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	// cast ID as ObjectID
	objectID := bson.ObjectIdHex(ID)

	// delete record
	qry := bson.M{"_id": objectID}
	err := coll.Remove(qry)
	if err != nil {
		return err
	}

	return nil
}

// EbookGetByID retrieves an ebook given its mongodb ID
func EbookGetByID(ID string) (Ebook, error) {
	ebk := Ebook{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	// cast ID as ObjectID
	objectID := bson.ObjectIdHex(ID)

	qry := bson.M{"_id": objectID}
	err := coll.Find(qry).One(&ebk)
	if err != nil {
		return ebk, err
	}

	return ebk, nil
}

// EbookGetByIsbns retrieves an ebook given a set of isbns
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

// EbookUpdate saves an ebk struct to DB
func EbookUpdate(ebk Ebook) (Ebook, error) {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getEbooksColl()

	// let's add the time and save
	ebk.DateUpdated = time.Now()

	// we select on the ebook's ID
	selector := bson.M{"_id": ebk.ID}

	err := coll.Update(selector, &ebk)
	if err != nil {
		logger.Error.Printf("Couldn't update ebook: %v", err)
		return ebk, err
	}

	return ebk, nil
}

// EbooksCount counts the number of ebooks in DB
func EbooksCount() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks
	qry := coll.Find(nil)
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// EbooksCountPPNs retrieves the number of record that have a PPN field
func EbooksCountPPNs() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks
	qry := coll.Find(bson.M{"ppns": bson.M{"$exists": true}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// EbooksCountUnimarc retrieves the number of record that have a RecordUnimarc field
func EbooksCountUnimarc() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks
	qry := coll.Find(bson.M{"recordunimarc": bson.M{"$exists": true}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
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
			ErrEbkCreate := EbookCreate(record)
			if ErrEbkCreate != nil {
				logger.Error.Println(ErrEbkCreate)
			}
			createdCounter++
			continue
		}

		// we did find the record
		// protect the fields in DB that exist but don't exist in updated struct,
		// i.e. created date, ppn, recordunimarc...
		// NOTE: probably better ways to do this, see omitempty in struct definition?
		// BUG: bug list of protected files is used for both xml and csv upload, but it should be diff.
		record.ID = existingRecord.ID
		record.DateCreated = existingRecord.DateCreated
		record.Ppns = existingRecord.Ppns
		record.RecordUnimarc = existingRecord.RecordUnimarc
		record.SFXLastHarvest = existingRecord.SFXLastHarvest
		record.SFXID = existingRecord.SFXID
		record.RecordMarc21 = existingRecord.RecordMarc21

		// save to DB
		_, ErrEbkUpdate := EbookUpdate(record)
		if ErrEbkUpdate != nil {
			logger.Error.Println(ErrEbkUpdate)
		}
		updatedCounter++
	}
	return createdCounter, updatedCounter, nil
}

// EbooksGetByTSName retrieves the ebooks which have a given target service
// i.e. belong to a given package
func EbooksGetByTSName(tsname string) ([]Ebook, error) {
	var result []Ebook

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	// FIXME: iterate over slices of 1000 ebooks
	// maybe use channels?
	// or else retrieve all then calculate result / 1000 and just manage display of chuncks of 1000 ebooks
	iter := coll.Find(bson.M{"targetservice.tsname": tsname}).Limit(1000).Iter()
	err := iter.All(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// EbooksGetNoPPNByTSName retrieves all ebooks with conditions : no PPN, given TS
// used to prepare query to sudoc isbn2ppn web service
func EbooksGetNoPPNByTSName(tsname string) ([]Ebook, error) {
	var result []Ebook

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct) and checks if PPN exists
	err := coll.Find(bson.M{"targetservice.tsname": tsname, "ppns": bson.M{"$exists": false}}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}

// EbooksGetWithPPNByTSName retrieves all ebooks with condition : has PPN, given TS
// used to prepare query to sudoc get record web service
func EbooksGetWithPPNByTSName(tsname string) ([]Ebook, error) {
	var result []Ebook

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct) and checks if PPN exists
	err := coll.Find(bson.M{"targetservice.tsname": tsname, "ppns": bson.M{"$exists": true}}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}

// EbooksGetWithUnimarcByTSName retrieves all ebooks with condition : has Unimarc Record, given TS
func EbooksGetWithUnimarcByTSName(tsname string) ([]Ebook, error) {
	var result []Ebook

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct) and checks if PPN exists
	err := coll.Find(bson.M{"targetservice.tsname": tsname, "recordunimarc": bson.M{"$exists": true}}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}
