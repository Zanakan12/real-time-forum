package db

import (
	"database/sql"
	"fmt"
)

func createNotificationsTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    comment_id INTEGER,
    like_dislike_id INTEGER,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (like_dislike_id) REFERENCES likes_dislikes(id) ON DELETE CASCADE,
    CHECK ((comment_id IS NULL AND like_dislike_id IS NOT NULL) OR (comment_id IS NOT NULL AND like_dislike_id IS NULL))
);
`

	executeSQL(db, createTableSQL)
}

func NotificationsSelect(userID int) ([]Notification, error) {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	query := `
		SELECT id, user_id, comment_id, like_dislike_id, is_read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := tx.Query(query, userID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error querying notifications: %v", err)
	}
	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.CommentID, &n.LikeDislikeID, &n.IsRead, &n.CreatedAt)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error scanning notification row: %v", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error iterating notification rows: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return notifications, nil
}

// UpdateComment updates an existing comment in the database.
func NotificationsUpdateIsRead(notificationID int) error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	updateSQL := `UPDATE notifications SET is_read = ? WHERE id = ?`
	_, err = tx.Exec(updateSQL, true, notificationID)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
