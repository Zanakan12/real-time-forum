package db

import (
	"database/sql"
	"fmt"
)

func createRequestsTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS requests (
		user_id INTEGER NOT NULL UNIQUE,
		user_username TEXT NOT NULL
	);`

	executeSQL(db, createTableSQL)
}


func RequestInsert(userID int, username string) error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	insertSQL := `INSERT INTO requests (user_id, user_username) VALUES (?, ?)`
	_, err = tx.Exec(insertSQL, userID, username)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}


func createRequestToAdminTable(db *sql.DB) {
	// SQL statement to create the users table if it does not already exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS moderator_request (
        post_id INTEGER UNIQUE,
        title VARCHAR(25) NOT NULL ,
        moderator_request VARCHAR(10) NOT NULL,
        admin_response VARCHAR(10) NULL
    );`
	executeSQL(db, createTableSQL)
}

func RequestToAdmin(postID int, title, moderatorRequest string, adminResponse sql.NullString) error {
	// Connect to the database
	db := SetupDatabase()
	defer db.Close()

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	var query string
	if adminResponse.Valid {
		// If there's an admin response, update the existing record
		query = `UPDATE moderator_request 
                 SET admin_response = ? 
                 WHERE post_id = ?`
		_, err := tx.Exec(query, adminResponse.String, postID)
		if err != nil {
			tx.Rollback() // Rollback the transaction on error
			return fmt.Errorf("error updating admin response: %v", err)
		}
	} else {
		// If no admin response, insert a new request into the table
		query = `INSERT OR IGNORE INTO moderator_request (post_id, title, moderator_request, admin_response) 
                 VALUES (?, ?, ?, NULL)`
		_, err := tx.Exec(query, postID, title, moderatorRequest)
		if err != nil {
			tx.Rollback() // Rollback the transaction on error
			return fmt.Errorf("error inserting moderator request: %v", err)
		}
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback if commit fails
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
