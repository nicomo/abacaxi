// Package models stores the structs for the objects we have & interacts with mongo
package models

import (
	"net/http"

	"github.com/microcosm-cc/bluemonday"
	"github.com/nicomo/abacaxi/logger"

	"gopkg.in/mgo.v2/bson"
)

// Search parses the search form in nav to retrieve ebooks
func Search(r *http.Request) ([]Record, string, error) {

	var results []Record

	// create sanitizing policy : strict
	p := bluemonday.StrictPolicy()

	// we parse the form
	err := r.ParseForm()
	if err != nil {
		logger.Error.Println(err)
		return results, "", err
	}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getRecordsColl()

	// build query
	qryString := p.Sanitize(r.FormValue("search_terms"))
	qry := bson.M{"$text": bson.M{"$search": qryString}}

	//TODO: sort by relevance. See https://docs.mongodb.com/manual/reference/operator/query/text/#sort-by-text-search-score
	// execute query
	ErrFind := coll.Find(qry).Limit(200).All(&results)
	if ErrFind != nil {
		return results, qryString, err
	}

	return results, qryString, nil

}
