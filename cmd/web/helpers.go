package main

import (
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/sheshan1961/chessapp/pkg/models"
	"net/http"
	"runtime/debug"
	"time"
)

//serveError helper that will write an error message and stack trace to the errorLog
//It will also send a generic 500 internal server error response to the user
func (app *application) serveError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

//clienterror helper will send an error to the client side
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

//generic 404 error helper
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serveError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	//sometimes, a page gets half-loaded with a runtime error, so we first write the template to a buffer we get from a buffer pool
	buf := app.sizedBufferPool.Get()

	//if there is an error we serve an error to the client instead of a half-loaded page
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serveError(w, err)
		return
	}
	buf.WriteTo(w)
	app.sizedBufferPool.Put(buf)
}

func (app *application) addDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}

	td.CurrentYear = time.Now().Year()
	//add flash message if one exists
	td.Flash = app.sessionManager.PopString(r.Context(), "flash")
	//add the value for userID (nil if none)
	td.AuthenticatedUser = app.authenticatedUser(r)
	//get the csrf token and pass it to the template data
	td.CSRFToken = nosurf.Token(r)
	return td
}

func (app *application) authenticatedUser(r *http.Request) *models.User {
	if user, ok := r.Context().Value(contextKeyUser).(*models.User); !ok {
		return nil
	} else {
		return user
	}
}
