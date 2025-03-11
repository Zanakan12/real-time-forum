package db

import (
	"database/sql"

	"fmt"
	"log"
)

func createMessagesTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
	recipient TEXT NOT NULL,
    content TEXT NOT NULL,
	read BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

`
	executeSQL(db, createTableSQL)
}

func GetMessages(username, recipient string) ([]WebSocketMessage, error) {
	db := SetupDatabase()
	defer db.Close()

	// ✅ Correction de la requête SQL pour récupérer les messages dans les deux sens
	query := `SELECT username, content, created_at
	          FROM messages 
	          WHERE (username = ? AND recipient = ?) 
	          OR (username = ? AND recipient = ?) 
	          ORDER BY created_at ASC`
	rows, err := db.Query(query, username, recipient, recipient, username)
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

// Stocke un message en base de données
func SaveMessage(username, recipient, content, created_at string, read bool) error {
	db := SetupDatabase()
	defer db.Close()
	_, err := db.Exec(`INSERT INTO messages (username, recipient, content, read, created_at) 
	                   VALUES (?, ?, ?, ?, ?)`, username, recipient, content, read, created_at)
	return err
}

// Récupère les messages non lus pour un utilisateur
func GetUnreadMessages(username string) []WebSocketMessage {
	db := SetupDatabase()
	defer db.Close()
	rows, err := db.Query(`SELECT username, recipient, content, created_at FROM messages 
	                       WHERE recipient = ? AND read = false ORDER BY created_at ASC`, username)
	if err != nil {
		log.Println("Erreur récupération messages non lus :", err)
		return nil
	}
	defer rows.Close()

	var messages []WebSocketMessage
	for rows.Next() {
		var msg WebSocketMessage
		err := rows.Scan(&msg.Username, &msg.Recipient, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Println("Erreur scan message non lu :", err)
			continue
		}
		msg.Read = false
		messages = append(messages, msg)
	}
	return messages
}

// Marque un message comme lu après envoi
func MarkMessageAsRead(msg WebSocketMessage) error {
	db := SetupDatabase()
	defer db.Close()
	_, err := db.Exec(`UPDATE messages SET read = true WHERE username = ? AND recipient = ? AND content = ?`,
		msg.Username, msg.Recipient, msg.Content)
	return err
}

func GetAllUser(aux []string) ([]User, error) {
	db := SetupDatabase()
	defer db.Close()

	// Requête SQL pour récupérer tous les utilisateurs
	query := "SELECT username FROM users"

	// Exécuter la requête
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'exécution de la requête : %v", err)
	}
	defer rows.Close()

	// Stocker les utilisateurs
	var users []User

	// Convertir aux en map pour une recherche rapide
	auxMap := make(map[string]bool)
	for _, name := range aux {
		auxMap[name] = true
	}

	// Parcourir les résultats de la requête SQL
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username); err != nil {
			return nil, fmt.Errorf("erreur lors de l'analyse des données : %v", err)
		}

		decryptedUser, _ := DecryptData(user.Username)

		// Ajouter l'utilisateur SEULEMENT s'il n'est pas dans auxMap

		if !auxMap[decryptedUser] {
			users = append(users, user)
		}
	}

	// Vérifier les erreurs d'itération
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération : %v", err)
	}

	return users, nil
}
