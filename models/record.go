package models

import (
	"time"

	"github.com/nicomo/abacaxi/logger"

	"gopkg.in/mgo.v2/bson"
)

const (
	IDTypeOnline = iota // Types of Identifiers: mostly online ISBN / ISSN
	IDTypePrint         // Types of Identifiers: mostly print ISBN / ISSN
	IDTypePPN           // Types of Identifiers: unimarc record ID in Sudoc catalog
	IDTypeSFX           // Types of Identifiers: ID in Ex Libris' SFX Open resolver
)

// Record stores a full record for a resource
type Record struct {
	ID                           bson.ObjectId `bson:"_id,omitempty"`
	AccessType                   string        `bson:",omitempty"`
	Acquired                     bool          `bson:",omitempty"`
	Active                       bool
	CoverageDepth                string `bson:",omitempty"`
	CoverageNotes                string `bson:",omitempty"`
	DateCreated                  time.Time
	DateFirstIssueOnline         string    `bson:",omitempty"`
	DateLastIssueOnline          string    `bson:",omitempty"`
	DateMonographPublishedOnline string    `bson:",omitempty"`
	DateMonographPublishedPrint  string    `bson:",omitempty"`
	DateUpdated                  time.Time `bson:",omitempty"`
	Deleted                      bool
	EmbargoInfo                  string       `bson:",omitempty"`
	FirstAuthor                  string       `bson:",omitempty"`
	FirstEditor                  string       `bson:",omitempty"`
	Identifiers                  []Identifier `bson:",omitempty"`
	MonographEdition             string       `bson:",omitempty"`
	MonographVolume              string       `bson:",omitempty"`
	Notes                        string       `bson:",omitempty"`
	NumFirstIssueOnline          string       `bson:",omitempty"`
	NumFirstVolOnline            string       `bson:",omitempty"`
	NumLastIssueOnline           string       `bson:",omitempty"`
	NumLastVolOnline             string       `bson:",omitempty"`
	ParentPublicationTitleID     string       `bson:",omitempty"`
	PrecedingPublicationTitleID  string       `bson:",omitempty"`
	PublicationTitle             string
	PublicationType              string    `bson:",omitempty"`
	PublisherName                string    `bson:",omitempty"`
	RecordMarc21                 string    `bson:",omitempty"`
	RecordUnimarc                string    `bson:",omitempty"`
	TargetServices               []TSEmbed `bson:",omitempty"` // this is the name of the package in SFX, e.g. CAIRN QSJ
	TitleID                      string    `bson:",omitempty"`
	TitleURL                     string    `bson:",omitempty"`
}

// Identifier embedded in an record
type Identifier struct {
	Identifier string `bson:",omitempty"`
	IDType     int
}

// TSEmbed embeds the necessary fields from a Target Service
type TSEmbed struct {
	Name        string
	DisplayName string
}

// RecordDelete deletes a single ebook from DB
func RecordDelete(ID string) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection records
	coll := getRecordsColl()

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

// RecordGetByID retrieves a record given its mongodb ID
func RecordGetByID(ID string) (Record, error) {
	record := Record{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection records
	coll := getRecordsColl()

	// cast ID as ObjectID
	objectID := bson.ObjectIdHex(ID)

	qry := bson.M{"_id": objectID}
	err := coll.Find(qry).One(&record)
	if err != nil {
		return record, err
	}

	return record, nil
}

// RecordUpdate saves an updated record struct to DB
func RecordUpdate(record Record) (Record, error) {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getRecordsColl()

	// let's add the time and save
	record.DateUpdated = time.Now()

	// we select on the record's ID
	selector := bson.M{"_id": record.ID}

	err := coll.Update(selector, &record)
	if err != nil {
		logger.Error.Printf("Couldn't update record: %v", err)
		return record, err
	}

	return record, nil
}

// RecordUpsert inserts or updates a record in DB
func RecordUpsert(record Record) (int, int, error) {

	var updated, upserted int

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getRecordsColl()

	// selectorQry
	var IDsToQry []bson.M
	for i := 0; i < len(record.Identifiers); i++ {
		sTerm := bson.M{"identifiers.identifier": record.Identifiers[i].Identifier}
		IDsToQry = append(IDsToQry, sTerm)
	}
	selectorQry := bson.M{
		"$or": IDsToQry,
	}

	// updateQry
	changeInfo, err := coll.Upsert(selectorQry, record)
	if err != nil {
		return updated, upserted, err
	}

	// changeInfo tells us if there's been an update or an insert
	if changeInfo.UpsertedId != nil {
		upserted++
	}
	if changeInfo.Updated != 0 {
		updated++
	}

	return updated, upserted, nil
}

func recordToKbart(record Record) []string {
	var printID, onlineID string

	for _, v := range record.Identifiers {
		if v.IDType == IDTypePrint {
			printID = v.Identifier
			continue
		}
		if v.IDType == IDTypeOnline {
			onlineID = v.Identifier
			continue
		}
		break
	}

	result := []string{
		record.PublicationTitle,
		printID,
		onlineID,
		record.DateFirstIssueOnline,
		record.NumFirstIssueOnline,
		record.NumFirstVolOnline,
		record.DateLastIssueOnline,
		record.NumLastVolOnline,
		record.NumLastIssueOnline,
		record.TitleURL,
		record.FirstAuthor,
		record.TitleID,
		record.EmbargoInfo,
		record.CoverageDepth,
		record.CoverageNotes,
		record.PublisherName,
	}

	return result

}

// RecordsCount counts the number of records in DB
func RecordsCount() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks
	qry := coll.Find(nil)
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// RecordsCountPPNs retrieves the number of record that have a PPN Identifier
func RecordsCountPPNs() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks
	qry := coll.Find(bson.M{"identifiers.idtype": IDTypePPN})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// RecordsCountUnimarc retrieves the number of record that have a RecordUnimarc field
func RecordsCountUnimarc() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks
	qry := coll.Find(bson.M{"recordunimarc": bson.M{"$exists": true}})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// RecordsGetByTSName retrieves the records which have a given target service
// i.e. belong to a given package.
// n is used to paginate. Use 0 if you want to start at record #1
func RecordsGetByTSName(tsname string, n int) ([]Record, error) {
	var result []Record

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	q := coll.Find(bson.M{"targetservices.name": tsname}).Sort("publicationtitle")

	// skip to result number n
	// NOTE: if we want to paginate on large sets, we shouldn't skip
	// (mongo iterates over all the result documents and omits the first n that need to be skipped.)
	// better solution - see https://github.com/icza/minquery
	if n > 0 {
		q = q.Skip(n)
	}

	if err := q.All(&result); err != nil {
		return result, err
	}

	return result, nil
}

// RecordsGetNoPPNByTSName retrieves all records with conditions : no PPN, given TS
// used to prepare query to sudoc isbn2ppn web service
func RecordsGetNoPPNByTSName(tsname string) ([]Record, error) {
	var result []Record

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Record struct) and checks that PPN does not exist
	err := coll.Find(bson.M{"targetservices.name": tsname, "identifiers.idtype": bson.M{"$ne": IDTypePPN}}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}

// RecordsGetWithPPNByTSName retrieves all records with condition : has PPN, given TS
// used to prepare query to sudoc get record web service
func RecordsGetWithPPNByTSName(tsname string) ([]Record, error) {
	var result []Record

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Record struct) and checks if PPN exists
	err := coll.Find(bson.M{"targetservices.name": tsname, "identifiers.idtype": IDTypePPN}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}

// RecordsGetWithUnimarcByTSName retrieves all records with condition : has Unimarc Record, given TS
func RecordsGetWithUnimarcByTSName(tsname string) ([]Record, error) {
	var result []Record

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getRecordsColl()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Record struct) and checks if PPN exists
	err := coll.Find(bson.M{"targetservices.name": tsname, "recordunimarc": bson.M{"$exists": true}}).All(&result)
	if err != nil {
		logger.Error.Println(err)
		return result, err
	}

	return result, nil
}

// RecordsUpsert updates or inserts a number of records in DB
func RecordsUpsert(records []Record) (int, int) {
	var recordsUpdates, recordsInserts int
	for _, r := range records {
		updated, upserted, err := RecordUpsert(r)
		if err != nil {
			logger.Error.Println(err)
		}
		recordsUpdates += updated
		recordsInserts += upserted
	}
	return recordsUpdates, recordsInserts
}
