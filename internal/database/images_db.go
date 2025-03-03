package db

import (
	"config"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
)

func createImagesTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
`

	executeSQL(db, createTableSQL)
}

func ImageInsert(postId, fileSize int, filePath string) error {
	// Verify if the file has an extension.
	fileExt := strings.ToLower(filepath.Ext(filePath))
	if fileExt == "" {
		return fmt.Errorf("the file doesn't have an extension")
	}

	// Verify if the file have a valid extension.
	validExtension := false
	for _, ext := range config.IMG_EXT {
		if ext == fileExt {
			validExtension = true
			break
		}
	}

	if !validExtension {
		return fmt.Errorf("the %s extension is not valid", fileExt)
	}

	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	encryptedFilePath, err := encryptData(filePath)
	if err != nil {
		return fmt.Errorf("error encrypting title: %v", err)
	}

	createSQL := `INSERT INTO images (post_id, file_path, file_size) VALUES (?, ?, ?)`
	_, err = tx.Exec(createSQL, postId, encryptedFilePath, fileSize)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

// GetImagesByPostID retrieves all image records for a specific post ID.
func ImageSelectByPostID(postID int, db *sql.DB) ([]Images, error) {

	// Begin a new transaction.
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Ensure that the transaction is rolled back in case of an error.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the SQL query.
	query := `SELECT id, post_id, file_path, file_size, created_at FROM images WHERE post_id = ?`

	// Execute the query within the transaction.
	rows, err := tx.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var images []Images
	for rows.Next() {
		var img Images
		if err := rows.Scan(&img.ID, &img.PostID, &img.FilePath, &img.FileSize, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		images = append(images, img)
	}

	// Check for errors during iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return images, nil
}

// This function manages its own database connection and uses transactions.
func ImageDeleteByPostID(postID int64) error {

	db := SetupDatabase()
	defer db.Close()

	// Start a new transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Ensure that the transaction is rolled back in case of an error or panic.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the DELETE SQL query.
	query := `DELETE FROM images WHERE post_id = ?`

	// Execute the query within the transaction.
	_, err = tx.Exec(query, postID)
	if err != nil {
		return fmt.Errorf("error executing delete query: %v", err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
