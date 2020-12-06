package main

import (
	"context"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/sheshan1961/chessapp/pkg/models"
	"net/http"
)

//Set some common http headers to all of our handlers
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

//log each request made by getting the IP, the protocol version, the method and the query
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serveError(w, fmt.Errorf("%e", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.authenticatedUser(r) == nil {
			http.Redirect(w, r, "/user/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check if userID exists in the database
		//if it's not present, we just move on to the next handler
		exists := app.sessionManager.Exists(r.Context(), "userID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}
		//fetch the details of the user and if there's not a match, we remove it from the session manager and call next handler
		//if it's another error, we return the error to the user
		user, err := app.users.Get(app.sessionManager.GetInt(r.Context(), "userID"))
		if err == models.ErrNoRecord {
			app.sessionManager.Remove(r.Context(), "userID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serveError(w, err)
			return
		}

		//create a new copy of the request with the authenticated user added to the context and then call next handler with new request
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}
