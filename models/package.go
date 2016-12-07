// Package models stores the structs for the objects we have
package models

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

func PackageCountEbooks(packname string) int {
	logger.Debug.Println("PackageCountEbooks: ", packname)
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": packname})
	logger.Debug.Println(qry)
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}
