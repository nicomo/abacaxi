package controllers

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/models"
	"github.com/nicomo/abacaxi/session"
	"github.com/nicomo/abacaxi/views"
)

const (
	// sesssion name to store the number of login attemps
	sessLoginAttempt = "log_attempts"
)

// logAttempt increments a counter of logging attempts in session
func logAttempt(sess *sessions.Session) {
	// log the login attempt
	if sess.Values[sessLoginAttempt] == nil {
		sess.Values[sessLoginAttempt] = 1
	} else {
		sess.Values[sessLoginAttempt] = sess.Values[sessLoginAttempt].(int) + 1
	}
}

// UsersHandler displays the list of existing users
func UsersHandler(w http.ResponseWriter, r *http.Request) {}

// UsersLoginGetHandler
func UsersLoginGetHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTmpl(w, "userlogin", nil)
}

// UsersLoginPostHandler
func UsersLoginPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r, "abacaxi-session")

	logger.Debug.Println(sess)

	// Prevent brute force login attempt: invalidate request
	if sess.Values[sessLoginAttempt] != nil && sess.Values[sessLoginAttempt].(int) >= 5 {
		logger.Info.Println("Brute force login prevented")
		sess.Save(r, w)
		UsersLoginGetHandler(w, r)
		return
	}

	// new strict sanitizing policy for the login form
	policy := bluemonday.StrictPolicy()
	// get form values
	username := policy.Sanitize(r.FormValue("username"))
	pw := policy.Sanitize(r.FormValue("password"))

	logger.Debug.Println(pw)

	if username == "" || pw == "" {
		logger.Info.Println("login attempt missing required field")
		sess.Save(r, w)
		UsersLoginGetHandler(w, r)
		return
	}

	// Get user in DB
	user, err := models.UserByUsername(username)
	if err != nil {
		logger.Error.Println(err)
		logAttempt(sess)
		UsersLoginGetHandler(w, r)
		return
	}

	if session.MatchString(user.Password, pw) {
		// login is successful

		// clean session (of login attempts counter)
		delete(sess.Values, sessLoginAttempt)

		// fill session values, save & redirect to home
		sess.Values["id"] = user.ID.Hex()
		sess.Values["username"] = user.Username
		sess.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else {
		logger.Error.Println("wrong password")
		logAttempt(sess)
		UsersLoginGetHandler(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// UsersLogoutHandler logs user out
func UsersLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r, "abacaxi-session")

	// If user is authenticated we empty the session
	if sess.Values["id"] != nil {
		session.Empty(sess)
		sess.Save(r, w)
	}

	// now logged out, redirect to home
	http.Redirect(w, r, "/", http.StatusFound)
}

// UsersNewGetHandler
func UsersNewGetHandler(w http.ResponseWriter, r *http.Request) {}

// UsersNewPostHandler
func UsersNewPostHandler(w http.ResponseWriter, r *http.Request) {}

// UsersDeleteHandler
func UsersDeleteHandler(w http.ResponseWriter, r *http.Request) {}
