package db

import (
	"database/sql"
	"fmt"
)

func createMessagesTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
	recipient TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

`
	executeSQL(db, createTableSQL)
}

// SaveMessage enregistre un message dans la base de données
func SaveMessage(username, recipient, content string) error {
	db := SetupDatabase()
	defer db.Close()

	query := `INSERT INTO messages (username,recipient, content) VALUES (?, ?, ?)`
	_, err := db.Exec(query, username, recipient, content)
	if err != nil {
		return fmt.Errorf("Erreur lors de l'insertion du message : %v", err)
	}

	return nil
}

func GetMessages(username string) ([]WebSocketMessage, error) {

	db := SetupDatabase()
	defer db.Close()

	query := `SELECT username, content, created_at
	FROM messages 
	WHERE username = ? 
	ORDER BY created_at ASC`
	rows, err := db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []WebSocketMessage
	for rows.Next() {
		var msg WebSocketMessage
		err := rows.Scan(&msg.Username, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
