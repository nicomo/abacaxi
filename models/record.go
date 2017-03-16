package models

import (
	"time"

	"github.com/nicomo/abacaxi/logger"

	"gopkg.in/mgo.v2/bson"
)

const (
	IdTypeOnline = iota
	IdTypePrint
	IdTypePPN
	IdTypeSFX
)

type Record struct {
	ID                           bson.ObjectId `bson:"_id,omitempty"`
	AccessType                   string        `bson:",omitempty"`
	Acquired                     bool          `bson:",omitempty"`
	Active                       bool
	CoverageDepth                string `bson:",omitempty"`
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
	PublicationType              string          `bson:",omitempty"`
	PublisherName                string          `bson:",omitempty"`
	RecordMarc21                 string          `bson:",omitempty"`
	RecordUnimarc                string          `bson:",omitempty"`
	TargetServices               []TargetService `bson:",omitempty"` // this is the name of the package in SFX, e.g. CAIRN QSJ
	TitleID                      string          `bson:",omitempty"`
	TitleURL                     string          `bson:",omitempty"`
}

// Identifier embedded in an record
type Identifier struct {
	Identifier string
	IdType     int
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
		logger.Error.Printf("could not save record with identifier %v in DB: %s", record.Identifiers[0].Identifier, err)
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
