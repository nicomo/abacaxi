package models

import (
	"log"
	"time"

	"github.com/nicomo/EResourcesMetadataHub/logger"

	"gopkg.in/mgo.v2"
)

const (
	MongoDBHosts = "localhost:27017"
	AuthDatabase = "metadatahub"
)

var mgoSession *mgo.Session

// FIXME: a package level var is not the right way to maintena a mongo session
// see http://stackoverflow.com/questions/26574594/best-practice-to-maintain-a-mgo-session/26576589#26576589
// https://groups.google.com/forum/#!topic/golang-nuts/g_zHm1E3sIs
// http://stackoverflow.com/questions/37041430/is-there-a-standard-way-to-keep-a-database-session-open-across-packages-in-golan

func init() {

	// ifo required to get a session to mongoDB
	mgoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{MongoDBHosts},
		Timeout:  60 * time.Second,
		Database: AuthDatabase,
	}

	//  mgoSession maintains a pool of socket connections to mongoDB
	var err error
	mgoSession, err = mgo.DialWithInfo(mgoDBDialInfo)
	if err != nil {
		log.Fatalf("cannot dial mongodb: %s\n", err)
	}

	mgoSession.SetMode(mgo.Monotonic, true)

	// create the Target Services collection, with an index on names
	tsColl := mgoSession.DB(AuthDatabase).C("targetservices")
	tsIndex := mgo.Index{
		Key:        []string{"tsname"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	tsCollIndexErr := tsColl.EnsureIndex(tsIndex)
	if tsCollIndexErr != nil {
		panic(tsCollIndexErr)
	}

	// create the ebooks collection with a compound text index
	// see https://code.tutsplus.com/tutorials/full-text-search-in-mongodb--cms-24835
	ebkColl := mgoSession.DB(AuthDatabase).C("ebooks")
	ebkIndex := mgo.Index{
		Key:        []string{"$text:title", "$text:publisher", "$text:isbns.isbn", "$text:ppns.ppn"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}

	ebkIndexErr := ebkColl.EnsureIndex(ebkIndex)
	if ebkIndexErr != nil {
		logger.Error.Println(ebkIndexErr)
		//panic(ebkIndexErr)
	}

}

// getEbooksColl retrieves a pointer to the Ebooks mongo collection
func getEbooksColl() *mgo.Collection {
	ebksColl := mgoSession.DB(AuthDatabase).C("ebooks")
	return ebksColl
}

// getTargetServiceColl retrieves a pointer to the Target Services (i.e. ebook commercial packages) mongo collection
func getTargetServiceColl() *mgo.Collection {
	tsColl := mgoSession.DB(AuthDatabase).C("targetservices")
	return tsColl
}
