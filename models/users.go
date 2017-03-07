package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// User contains the info for each user
type User struct {
	ID           bson.ObjectId `bson:"_id"`
	DateCreated  time.Time
	DateLastSeen time.Time `bson:",omitempty"`
	Username     string    `bson:"username"`
	Password     string    `bson:"password"`
}

// UserByUsername retrieves a user by its username
func UserByUsername(username string) (User, error) {
	user := User{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection users
	coll := getUsersColl()

	qry := bson.M{"username": username}
	err := coll.Find(qry).One(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// UserCreate creates a new user
func UserCreate(username, password string) error {
	now := time.Now()

	user := &User{
		ID:           bson.NewObjectId(),
		Username:     username,
		Password:     password,
		DateCreated:  now,
		DateLastSeen: now,
	}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection users
	coll := getUsersColl()

	err := coll.Insert(user)
	if err != nil {
		return err
	}

	return nil
}

// UserDelete deletes a user
func UserDelete(username string) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getUsersColl()

	// delete record
	qry := bson.M{"username": username}
	err := coll.Remove(qry)
	if err != nil {
		return err
	}

	return nil
}
