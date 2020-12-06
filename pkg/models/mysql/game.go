package mysql

import (
	"database/sql"
	"github.com/sheshan1961/chessapp/pkg/models"
	"log"
)

type GameModel struct {
	DB *sql.DB
}

//This will be used for the homepage "create game"
func (m *GameModel) Insert(key, fen string) (string, error) {
	//statement to execute the insert for our table
	stmt := `INSERT INTO game(room, fen, canChange, expires) VALUES (?, ?, TRUE,DATE_ADD(UTC_TIMESTAMP, INTERVAL 7 DAY))`

	//DB.execute will run the statement and return an error if it could not insert it
	_, err := m.DB.Exec(stmt, key, fen)
	if err != nil {
		return string('0'), err
	}
	return key, nil
}

//This will be used for the homepage "get game with room key"
//Step 1: Get the DB
//Change the canChange Value to True
//In order to make sure the same connection is used, we will be use the transactions
func (m *GameModel) Get(key string) (*models.Game, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	//set the canChange Status to True
	stmt := `UPDATE game SET canChange = TRUE WHERE room = ?`
	_, err = tx.Exec(stmt, key)
	//If error occurred, rollback so we make no changes
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	//Statement to get the room data if it exists
	stmt = `SELECT room, fen, canChange, expires FROM game WHERE room = ?`

	//use the above statement to get our room
	row := tx.QueryRow(stmt, key)
	g := &models.Game{}
	err = row.Scan(&g.Key, &g.Fen, &g.CanChange, &g.Expires)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, models.ErrNoRecord
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}
	//If everything works, then we can commit and then return with our data
	err = tx.Commit()
	return g, err
}

//This will be used when the users' decide to save the game to resume later
//Check if canChange is true to make changes, otherwise, can't do anything
//Changes to make: canChange should be false, fen string should change and expire date increases by another seven days
func (m *GameModel) Save(key, fen string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	stmt := `UPDATE game SET canChange = False WHERE room = ?`
	_, err = tx.Exec(stmt, key)
	//If error occurred, rollback so we make no changes
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatalf("Unable to rollback from Update canChange: %e", rollbackErr)
		}
		log.Fatalf("Update canChange Failed: %e", err)
		return err
	}

	stmt = `UPDATE game SET fen = ? WHERE room = ?`
	_, err = tx.Exec(stmt, fen, key)
	//If error occurred, rollback so we make no changes
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatalf("Unable to rollback from Update fen: %e", rollbackErr)
		}
		log.Fatalf("Update fen Failed: %e", err)
		return err
	}

	stmt = `UPDATE game SET expires = DATE_ADD(UTC_TIME(), INTERVAL 7 DAY) WHERE room = ?`

	_, err = tx.Exec(stmt, key)
	//If error occurred, rollback so we make no changes
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatalf("Unable to rollback from Update expires: %e", rollbackErr)
		}
		log.Fatalf("Update expires Failed: %e", err)
		return err
	}

	//If everything works, then we can commit and then return with our data
	err = tx.Commit()
	return err
}
