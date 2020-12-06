package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sheshan1961/chessapp/cmd/websockets"
	"github.com/sheshan1961/chessapp/pkg/forms"
	"github.com/sheshan1961/chessapp/pkg/models"
	"net/http"
)

const startingFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	//This runs when a post request is made instead of a Get (basically when one of the two buttons are hit)
	if r.Method == "POST" {
		//must parse form before looking at value
		err := r.ParseForm()
		//in case something went wrong with our request
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		//get the submit value to see which of the two input pressed
		submit := r.PostForm.Get("submit")
		//switch statement to run a func depending on which input pressed
		switch submit {
		case "Create a Room":
			app.createRoom(w, r)
			return
		case "Enter Room":
			//check for empty string
			form := forms.New(r.PostForm)
			form.Required("room-code")
			if !form.Valid() {
				app.render(w, r, "home.page.tmpl", &TemplateData{Form: form})
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/game/%s", form.Get("room-code")), http.StatusSeeOther)
			return
		}

	}

	app.render(w, r, "home.page.tmpl", &TemplateData{Form: forms.New(nil)})
}

//This is the post form method for when the user hits on the "Create a Room" button
func (app *application) createRoom(w http.ResponseWriter, r *http.Request) {
	//generate a random uuid number
	key := uuid.New().String()
	//since this button will only be pressed when a user wants a new game, we need to pass in the fen key for a starting board
	fen := startingFen

	//Execute the insert command to made a new record of the room
	room, err := app.games.Insert(key, fen)
	//check if anything went wrong in executing the command
	if err != nil {
		app.serveError(w, err)
		return
	}
	//use Put() to add a string value of "Room Successfully Created. Invite your friend using the url"
	app.sessionManager.Put(r.Context(), "flash", "Room Successfully Created. Invite your friend using the url")
	//redirect the user to their newly created room
	http.Redirect(w, r, fmt.Sprintf("/game/%s", room), http.StatusSeeOther)
}

//This will be used to lead the player to the game. Will be called when "Enter room" is hit and after createRoom finishes making a new room
func (app *application) playGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	room := vars["key"]
	//check if the room is an empty string or not
	if room == "" {
		app.notFound(w)
		return
	}
	//check if room exists and if it does and an error still occurs, it's from server side
	g, err := app.games.Get(room)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serveError(w, err)
		return
	}

	//create templateData struct to hold all the data to pass through
	data := &TemplateData{Game: g, Hub: app.hub}
	//push the starting fen into hub if the length of it is zero
	if len(app.hub.MovesList[room]) == 0 {
		app.hub.MovesList[room] = append(app.hub.MovesList[room], g.Fen)
	}

	app.render(w, r, "game.page.tmpl", data)
}

//websocket handler
func (app *application) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := websockets.Upgrader.Upgrade(w, r, nil)
	vars := mux.Vars(r)
	room := vars["key"]
	if err != nil {
		app.serveError(w, err)
		return
	}

	//make a new subscription and connection
	connection := &websockets.Connection{Ws: conn, Send: make(chan *websockets.JsonInfo, 256)}
	subscription := &websockets.Subscription{Connection: connection, Hub: app.hub, Room: room}

	//register the client
	app.hub.Register <- subscription
	//start goroutines to read and write messages by the client
	go subscription.ReadPump()
	go subscription.WritePump()
}

/**
USER AUTH
*/

//The sign up form displayed to the user
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &TemplateData{Form: forms.New(nil)})
}

//Validate the values passed in for the signup and add the new user to the database
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	//check for any errors
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &TemplateData{Form: form})
		return
	}

	//make a new user based on the provided data
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	//do error checking and if an email is already made using the user-specified email, send the error back to the user
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.tmpl", &TemplateData{Form: form})
		return
	} else if err != nil {
		app.serveError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

//display a new form for the users to sign in

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &TemplateData{Form: forms.New(nil)})
}

//validate the login of the user
//add the user id of the signed in user to the session
//renew the session token to prevent session fixation issues
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	//check for correct credentials
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &TemplateData{Form: form})
		return
	} else if err != nil {
		app.serveError(w, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serveError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//logout the user by removing the session Data
//send a flash message to the session to give confirmation to user
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serveError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "userID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
