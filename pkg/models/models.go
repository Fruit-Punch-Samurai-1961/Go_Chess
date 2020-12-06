package models

import (
	"errors"
	"time"
)

//custom errors to send to user when we interact with the database and something goes wrong
var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

type Game struct {
	Key       string
	Fen       string
	CanChange bool
	Expires   time.Time
}

type User struct {
	ID int
	Name string
	Email string
	HashedPassword []byte
	Created time.Time
}
