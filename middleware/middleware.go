package middleware

import (
	"net/http"

	"github.com/nicomo/abacaxi/session"
)

// DisallowAnon does not allow anonymous users to access the page
func DisallowAnon(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session
		sess := session.Instance(r)

		// If user is not authenticated, redirect to login
		if sess.Values["id"] == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// otherwise, move on with context logged in true
		h.ServeHTTP(w, r)
	})
}

// DisallowAuthed prevents logged in users to access /users/login
func DisallowAuthed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session
		sess := session.Instance(r)

		// If user is authenticated, redirect to home
		if sess.Values["id"] != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// otherwise, move on with context logged in true
		h.ServeHTTP(w, r)
	})
}
