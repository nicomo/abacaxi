package models

import (
	"gopkg.in/mgo.v2"
)

// getTargetServiceColl retrieves a pointer to the Target Services (i.e. ebook commercial packages) mongo collection
func getTargetServiceColl() *mgo.Collection {
	tsColl := mgoSession.DB(conf.AuthDatabase).C("targetservices")
	return tsColl
}

func getUsersColl() *mgo.Collection {
	usersColl := mgoSession.DB(conf.AuthDatabase).C("users")
	return usersColl
}

func getRecordsColl() *mgo.Collection {
	recordsColl := mgoSession.DB(conf.AuthDatabase).C("records")
	return recordsColl
}

func getReportsColl() *mgo.Collection {
	reportsColl := mgoSession.DB(conf.AuthDatabase).C("reports")
	return reportsColl
}
