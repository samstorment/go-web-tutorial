package middleware

import (
	"net/http"
	"../sessions"
)


// Takes a http handlerFunc as an argument, performs the middle ware roles THEN calls the argument function
func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the session for the request
		session, _ := sessions.Store.Get(r, "session")
		// get the username for the session cookies storage
		_, ok := session.Values["user_id"]
		// if the username is bad, redirect to the login page
		if !ok { 
			http.Redirect(w, r, "/login", 302)
			// if this return is hit, the argument handler is never called, which is good
			return
		}
		// THIS line calls the HandlerFunc passed as an argument
		handler.ServeHTTP(w, r)
	}
}