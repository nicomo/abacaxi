package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	IdTypeOnline = iota
	IdTypePrint
	IdTypePPN
	IdTypeSFX
)

type Record struct {
	ID                           bson.ObjectId `bson:"_id"`
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
	EmbargoInfo                  string `bson:",omitempty"`
	FirstAuthor                  string `bson:",omitempty"`
	FirstEditor                  string `bson:",omitempty"`
	Identifiers                  []Identifier `bson:",omitempty"`
	MonographEdition             string `bson:",omitempty"`
	MonographVolume              string `bson:",omitempty"`
	Notes                        string `bson:",omitempty"`
	NumFirstIssueOnline          string `bson:",omitempty"`
	NumFirstVolOnline            string `bson:",omitempty"`
	NumLastIssueOnline           string `bson:",omitempty"`
	NumLastVolOnline             string `bson:",omitempty"`
	ParentPublicationTitleID     string `bson:",omitempty"`
	PrecedingPublicationTitleID  string `bson:",omitempty"`
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
func RecordsUpsert(records []Record) {}


func RecordUpsert(record Record) error {

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
	info, err := coll.Upsert(selectorQry, record)
	logger.Debug.Println(info)
	if err != nil {
		logger.Error.Printf("could not save record with identifier %v in DB: %s", record.Identifiers[0].Identifier, err)
		return err
	}

	return nil


}









}