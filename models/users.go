package models

import (
	"time"

	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/session"

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

// GetUsers retrieves the full list of users
func GetUsers() ([]User, error) {

	var Users []User

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getUsersColl()

	err := coll.Find(bson.M{}).Sort("username").All(&Users)
	if err != nil {
		return Users, err
	}

	return Users, nil
}

func UserByID(ID string) (User, error) {
	user := User{}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection users
	coll := getUsersColl()

	// cast ID as ObjectID
	objectID := bson.ObjectIdHex(ID)

	// delete record
	qry := bson.M{"_id": objectID}
	err := coll.Find(qry).One(&user)
	if err != nil {
		return user, err
	}

	return user, nil
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
	// hashing the password
	pw, err := session.HashString(password)
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	user := &User{
		ID:          bson.NewObjectId(),
		Username:    username,
		Password:    pw,
		DateCreated: now,
	}

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection users
	coll := getUsersColl()

	errColl := coll.Insert(user)
	if errColl != nil {
		return errColl
	}

	return nil
}

// UserDelete deletes a user
func UserDelete(ID string) error {

	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection ebooks
	coll := getUsersColl()

	// cast ID as ObjectID
	objectID := bson.ObjectIdHex(ID)

	// delete record
	qry := bson.M{"_id": objectID}
	err := coll.Remove(qry)
	if err != nil {
		return err
	}

	return nil
}

// UserUpdateDateLastSeen updates a user's record when she logs in
func UserUpdateDateLastSeen(u User) error {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()

	// collection users
	coll := getUsersColl()

	// update query
	qry := bson.M{ // field to update
		"$set": bson.M{"datelastseen": u.DateLastSeen},
	}
	err := coll.Update(bson.M{"_id": u.ID}, qry)
	if err != nil {
		return err
	}

	return nil
}

func UsersCount() int {
	// Request a socket connection from the session to process our query.
	mgoSession := mgoSession.Copy()
	defer mgoSession.Close()
	coll := getUsersColl()

	qry := coll.Find(bson.M{})
	count, err := qry.Count()
	if err != nil {
		return 0
	}

	return count
}
