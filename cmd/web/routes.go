package main

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"net/http"
)

//function for all of our route declarations
func (app *application) routes(cfg *Config) http.Handler {
	//make a variable containing all of our middleware
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	//middleware that goes to only certain routes
	dynamicMiddleware := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	r := mux.NewRouter()
	r.Handle("/", dynamicMiddleware.ThenFunc(app.home)).Methods("GET", "POST")
	r.Handle("/game/{key}", dynamicMiddleware.ThenFunc(app.playGame))
	r.Handle("/game/{key}/wss", dynamicMiddleware.ThenFunc(app.serveWs))

	//User auth routes
	r.Handle("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm)).Methods("GET")
	r.Handle("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser)).Methods("POST")
	r.Handle("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm)).Methods("GET")
	r.Handle("/user/login", dynamicMiddleware.ThenFunc(app.loginUser)).Methods("POST")
	r.Handle("/user/logout", dynamicMiddleware.ThenFunc(app.logoutUser)).Methods("POST")

	//handle static files -> make it like file path so in our case to get it we need to go to ui then to static
	r.PathPrefix("/ui/static/").Handler(http.StripPrefix("/ui/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	return standardMiddleware.Then(r)
}
