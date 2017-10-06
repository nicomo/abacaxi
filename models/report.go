package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	UploadCsv   = iota // Types of batch operation: publisher csv upload
	UploadKbart        // Types of batch operation: kbart csv upload
	UploadSfx          // Types of batch operation: sfx xml upload
	SudocWs            // Types of batch operation: retrieve Unimarc Records from Sudoc Web Service
)

// Report is a report about a batch operation, stored in DB
type Report struct {
	ID          bson.ObjectId `bson:"_id"`
	DateCreated time.Time
	ReportType  int
	Text        []string
	Success     bool
}

// ReportsGet retrieves the last 100 reports from the DB
func ReportsGet() ([]Report, error) {
	var Reports []Report

	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getReportsColl()

	// we only need the last 100 reports
	q := coll.Find(nil).Sort("-datecreated").Limit(100)
	if err := q.All(&Reports); err != nil {
		return Reports, err
	}
	return Reports, nil
}

// ReportCreate inserts a report into the DB
func (report *Report) ReportCreate() error {
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getReportsColl()
	report.ID = bson.NewObjectId()
	report.DateCreated = time.Now()
	if err := coll.Insert(report); err != nil {
		return err
	}
	return nil
}
