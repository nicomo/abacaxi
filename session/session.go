package session

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/sessions"
)

var (
	// Store is the cookie store
	Store *sessions.CookieStore
)

func StoreCreate(ssk string) {
	Store = sessions.NewCookieStore([]byte(ssk))
}

// Instance returns a new session, never returns an error
func Instance(r *http.Request) *sessions.Session {
	sess, _ := Store.Get(r, "abacaxi-session")
	return sess
}

// Empty deletes all the current session values
func Empty(sess *sessions.Session) {
	// Clear out all stored values in the cookie
	for k := range sess.Values {
		delete(sess.Values, k)
	}
}

// HashString returns a hashed password string + error
func HashString(pw string) (string, error) {
	key, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(key), nil
}

// MatchString returns true if the hash matches the password
func MatchString(hash, pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	if err == nil {
		return true
	}

	return false
}
