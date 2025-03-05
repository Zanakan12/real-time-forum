package db

import (
	"database/sql"
	"fmt"
)

func createMessagesTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

`
	executeSQL(db, createTableSQL)
}

// SaveMessage enregistre un message dans la base de données
func  SaveMessage(msg Message) error {
	db := SetupDatabase()
	defer db.Close()

	query := `INSERT INTO messages (username, content, created_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, msg.Username, msg.Content, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("Erreur lors de l'insertion du message : %v", err)
	}
	return nil
}

func GetMessages() ([]Message, error) {

	db := SetupDatabase()
	defer db.Close()

	query := `SELECT id, username, content, created_at FROM messages ORDER BY created_at ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Username, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
