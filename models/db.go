package models

import (
	"log"
	"time"

	"github.com/nicomo/abacaxi/config"
	"github.com/nicomo/abacaxi/logger"

	"gopkg.in/mgo.v2"
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
		Timeout:  60 * time.Second,
		Database: conf.AuthDatabase,
	}

	//  mgoSession maintains a pool of socket connections to mongoDB
	var err error
	mgoSession, err = mgo.DialWithInfo(mgoDBDialInfo)
	if err != nil {
		log.Fatalf("cannot dial mongodb: %s\n", err)
	}

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

	ErrTSCollIndex := tsColl.EnsureIndex(tsIndex)
	if ErrTSCollIndex != nil {
		panic(ErrTSCollIndex)
	}

	// create the Users collection, with an index on username
	usersColl := mgoSession.DB(conf.AuthDatabase).C("users")
	usersIndex := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	ErrUsersCollIndex := usersColl.EnsureIndex(usersIndex)
	if ErrUsersCollIndex != nil {
		panic(ErrUsersCollIndex)
	}

	// create admin user if user collection is empty
	if UsersCount() == 0 {
		errUserOne := UserCreate("user1", "abacaxi-user1")
		if errUserOne != nil {
			logger.Error.Println(errUserOne)
		}
	}

	// create the records collection with a compound text index
	// see https://code.tutsplus.com/tutorials/full-text-search-in-mongodb--cms-24835
	recordsColl := mgoSession.DB(conf.AuthDatabase).C("records")
	recordIndex := mgo.Index{
		Key:        []string{"$text:publicationtitle", "$text:publishername", "$text:firstauthor", "$text:identifiers.identifier", "$text:ppns"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}

	ErrRecordIndex := recordsColl.EnsureIndex(recordIndex)
	if ErrRecordIndex != nil {
		logger.Error.Println(ErrRecordIndex)
	}

	recordIDIndex := mgo.Index{
		Key:        []string{"identifiers.identifier"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     false,
	}

	ErrRecordIDIndex := recordsColl.EnsureIndex(recordIDIndex)
	if ErrRecordIDIndex != nil {
		logger.Error.Println(ErrRecordIDIndex)
	}

}

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
