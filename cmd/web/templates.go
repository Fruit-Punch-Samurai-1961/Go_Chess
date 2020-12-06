package main

import (
	"github.com/sheshan1961/chessapp/cmd/websockets"
	"github.com/sheshan1961/chessapp/pkg/forms"
	"github.com/sheshan1961/chessapp/pkg/models"
	"html/template"
	"path/filepath"
)

type TemplateData struct {
	CurrentYear       int
	Form              *forms.Form
	Game              *models.Game
	Hub               *websockets.Hub
	Flash             string
	AuthenticatedUser *models.User
	CSRFToken         string
}

//func which will cache the templates for us so we don't end up calling the server so many times
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	//map to act as cache
	cache := map[string]*template.Template{}

	//grab all files that end in .page.tmpl into a slice
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	//loop through the pages
	for _, page := range pages {
		//get the filename
		name := filepath.Base(page)
		//parse the file and add into the initialized template set
		//If we wanted to add our own functions, then as we call it, we also need to send it our own functions
		//EX: template.New(name).Funcs(functions).ParseFiles(page)
		//Here, we create a new template with a name, name, give it our funcs, and then parse it
		//Create -> Add Own Funcs -> ParseFiles
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		//grab all the layouts and add them into the template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		//grab all the partials and add them into the template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
