package models

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MongoDBHosts = "localhost:27017"
	Database     = "metadatahub"
)

type Person struct {
	Name  string
	Phone string
}

func InitDB() {
	// connect to DB
	session, err := mgo.Dial(MongoDBHosts)
	if err != nil {
		log.Fatal("cannot dial mongodb", err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// collection ebooks
	c := session.DB(Database).C("ebooks")

	err = c.Insert(&Person{"Ale", "+55 53 8116 9639"},
		&Person{"Cla", "+55 53 8402 8510"})
	if err != nil {
		log.Fatal(err)
	}

	result := Person{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Phone)

}
