// Package models stores the structs for the objects we have
package models

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/nicomo/EResourcesMetadataHub/logger"
)

// PackageCountEbooks counts the number of ebooks for this package
func PackageCountEbooks(packname string) int {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": packname})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// PackageCountMarcRecords counts how many records for this package have proper MARC Records
func PackageCountMarcRecords(packname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": packname, "marcrecords": bson.M{"$ne": nil}})
	logger.Debug.Println(qry)
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}

// PackageCountPPNs counts how many records for this package have proper PicaPublication Numbers coming from ABES
func PackageCountPPNs(packname string) int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getEbooksCol()

	//  query ebooks by package name, aka Target Service in SFX (and in models.Ebook struct)
	qry := coll.Find(bson.M{"targetservice": packname, "ppns": false})
	count, err := qry.Count()

	if err != nil {
		logger.Error.Println(err)
	}

	return count
}
