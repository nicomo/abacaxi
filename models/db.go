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
	tsUsers := mgoSession.DB(conf.AuthDatabase).C("users")
	return tsUsers
}

func getRecordsColl() *mgo.Collection {
	recordsColl := mgoSession.DB(conf.AuthDatabase).C("records")
	return recordsColl
}
