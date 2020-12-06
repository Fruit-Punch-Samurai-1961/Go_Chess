package mysql

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/sheshan1961/chessapp/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type UserModel struct {
	DB *sql.DB
}

//Insert a new user into the database
func (m *UserModel) Insert(name, email, password string) error {
	//generate hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	//statement to enter the new record into the database
	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES (?, ?, ?, UTC_TIMESTAMP())`
	//execute the command
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "user_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
	}
	return err
}

//authenticate the user by looking up in the database if a email and password match
//We need to check for three things:
//1. Email exists in the database
//2. Password matches
//3. Return int
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	row := m.DB.QueryRow("SELECT id, hashed_password FROM users WHERE email = ?", email)
	err := row.Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}
	return id, nil
}

//Get details for specific user
func (m *UserModel) Get(id int) (*models.User, error) {
	s := &models.User{}
	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Name, &s.Email, &s.Created)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	return s, nil

}
