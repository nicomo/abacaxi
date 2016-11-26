package models

import (
	"log"

	"gopkg.in/mgo.v2"
)

const (
	MongoDBHosts = "localhost:27017"
	Database     = "metadatahub"
)

func getMgoSession() *mgo.Session {
	// connect to DB
	mgoSession, err := mgo.Dial(MongoDBHosts)
	if err != nil {
		log.Fatal("cannot dial mongodb", err)
	}

	mgoSession.SetMode(mgo.Monotonic, true)

	return mgoSession
}
