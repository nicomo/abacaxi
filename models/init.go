package models

import (
	"log"
	"time"

	"github.com/nicomo/abacaxi/config"
	"github.com/nicomo/abacaxi/logger"
	mgo "gopkg.in/mgo.v2"
)

// FIXME: a package level var is not the right way to maintena a mongo session
// see http://stackoverflow.com/questions/26574594/best-practice-to-maintain-a-mgo-session/26576589#26576589
// https://groups.google.com/forum/#!topic/golang-nuts/g_zHm1E3sIs
// http://stackoverflow.com/questions/37041430/is-there-a-standard-way-to-keep-a-database-session-open-across-packages-in-golan
// https://elithrar.github.io/article/custom-handlers-avoIDing-globals/
var mgoSession *mgo.Session
var conf config.Conf

func init() {

	// get the basic info for mongo
	conf = config.GetConfig()

	// info required to get a session to mongoDB
	mgoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.MongoDBHost},
		Timeout:  10 * time.Second,
		Database: conf.AuthDatabase,
	}

	//  mgoSession maintains a pool of socket connections to mongoDB
	logger.Info.Println("dialing mongodb...")
	var err error
	mgoSession, err = mgo.DialWithInfo(mgoDBDialInfo)
	if err != nil {
		logger.Error.Printf("DialWithInfo failed: %v", err)
		log.Fatalf("cannot dial mongodb: %s\n", err)
	}
	logger.Info.Println("... connection to mongodb OK")

	mgoSession.SetMode(mgo.Monotonic, true)

	// create the Target Services collection, with an index on names
	tsColl := mgoSession.DB(conf.AuthDatabase).C("targetservices")
	tsIndex := mgo.Index{
		Key:        []string{"tsname"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = tsColl.EnsureIndex(tsIndex)
	if err != nil {
		panic(err)
	}

	// create admin user if user collection is empty
	if UsersCount() == 0 {
		err = UserCreate("user1", "abacaxi-user1")
		if err != nil {
			logger.Error.Println(err)
		}
	}

	// create the records collection with a compound text index for general search
	// see https://code.tutsplus.com/tutorials/full-text-search-in-mongodb--cms-24835
	recordsColl := mgoSession.DB(conf.AuthDatabase).C("records")
	recordIndex := mgo.Index{
		Key:        []string{"$text:publicationtitle", "$text:publishername", "$text:firstauthor", "$text:identifiers.identifier", "$text:ppns"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}

	err = recordsColl.EnsureIndex(recordIndex)
	if err != nil {
		logger.Error.Println(err)
	}

	// create an index on records identifiers
	recordIDIndex := mgo.Index{
		Key:        []string{"identifiers.identifier"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}

	err = recordsColl.EnsureIndex(recordIDIndex)
	if err != nil {
		logger.Error.Println(err)
	}

	// create an ascending index on the field publicationtitle for recordsColl:
	recordTitleIndex := mgo.Index{
		Key:        []string{"publicationtitle"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}
	err = recordsColl.EnsureIndex(recordTitleIndex)
	if err != nil {
		logger.Error.Println(err)
	}
}
