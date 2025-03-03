package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func createPostsTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title VARCHAR(150) NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('published', 'draft', 'pending', 'irrelevant', 'obscene', 'illegal', 'insulting')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`
	executeSQL(db, createTableSQL)
}

func PostInsert(userID int, title, body string, categories []int) (int, error) {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}

	encryptedTitle, err := encryptData(title)
	if err != nil {
		return 0, fmt.Errorf("erreur lors du cryptage du titre: %v", err)
	}

	encryptedBody, err := encryptData(body)
	if err != nil {
		return 0, fmt.Errorf("erreur lors du cryptage du contenu: %v", err)
	}

	createSQL := `INSERT INTO posts (user_id, title, body, status) VALUES (?, ?, ?, ?)`
	result, err := tx.Exec(createSQL, userID, encryptedTitle, encryptedBody, "published")
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("erreur lors de l'exécution de la requête: %v", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("erreur lors de l'obtention de l'ID du dernier post inséré: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("erreur lors de l'engagement de la transaction: %v", err)
	}

	// Insérer les catégories
	categoriesInsertByPostId(postID, categories, db)

	return int(postID), nil // Retourner l'ID du post et nil pour l'erreur
}

func PostSelectByCategoryID(categoryID int64) ([]Post, error) {
	db := SetupDatabase()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("erreur lors du début de la transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	query := `
		SELECT p.id, p.user_id, p.title, p.body, p.status, p.created_at, p.updated_at
		FROM posts p
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
	`

	rows, err := tx.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête des posts par catégorie: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var encryptedTitle, encryptedBody string
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&encryptedTitle,
			&encryptedBody,
			&post.Status,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan des données du post: %v", err)
		}

		post.Title, err = DecryptData(encryptedTitle)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du décryptage du titre: %v", err)
		}

		post.Body, err = DecryptData(encryptedBody)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du décryptage du body: %v", err)
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération sur les posts: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("erreur lors du commit de la transaction: %v", err)
	}

	return posts, nil
}

func PostUpdateContent(id int, body string) error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	encryptedBody, err := encryptData(body)
	if err != nil {
		return fmt.Errorf("error decrypting body datas : %v", err)
	}

	updatedAt := time.Now()

	updateSQL := `UPDATE posts SET body=?, updated_at=? WHERE id=?`
	_, err = tx.Exec(updateSQL, encryptedBody, updatedAt, id)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func PostDelete(postID int) error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	deleteSQL := `DELETE FROM posts WHERE id=?`
	_, err = tx.Exec(deleteSQL, postID)

	if err != nil {
		tx.Rollback() // Rollback on error
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// CommentSelectByID retrieves a single comment for a specific comment ID from the database using a transaction
func PostTitleSelectById(postID int) (string, error) {
	db := SetupDatabase()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %v", err)
	}

	// Ensure rollback in case of an error or panic
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rethrow panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	query := `
        SELECT p.title
        FROM posts p
        WHERE p.id = ?`

	// Execute the query with the provided commentID
	var title string
	err = tx.QueryRow(query, postID).Scan(&title)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no comment found with ID %d", postID)
		}
		return "", fmt.Errorf("error executing query: %v", err)
	}

	// Decrypt the content
	decryptedTitle, decryptErr := DecryptData(title)
	if decryptErr != nil {
		return "", fmt.Errorf("error decrypting content for comment ID %d: %v", postID, decryptErr)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("error committing transaction: %v", err)
	}

	return decryptedTitle, nil
}

func UpdatePostStatus(id int, status string) error {
	// Setup the database connection
	db := SetupDatabase()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// SQL query to update only the status
	updateSQL := `UPDATE posts SET status=? WHERE id=?`

	// Execute the update statement with the status and id parameters
	_, err = tx.Exec(updateSQL, status, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// DisplaySignaledStatus returns a list of posts with statuses different from 'published', 'pending', and 'draft'.
func DisplaySignaledStatus() ([]Post, error) {
	// Setup the database connection
	db := SetupDatabase()
	defer db.Close()

	// SQL query to select posts with statuses other than 'published', 'pending', or 'draft'
	query := `SELECT id, title, status FROM posts WHERE status NOT IN ('published', 'pending', 'draft')`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %v", err)
	}
	defer rows.Close()

	// Slice to store the result posts
	var posts []Post

	// Iterate through the result rows and populate the posts slice
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Status)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		posts = append(posts, post)

	}

	// Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Return the slice of posts
	return posts, nil
}

// FetchModeratorRequests fetches all records from the moderator_request table.
func DisplayAdminResponse() ([]ModeratorRequest, error) {
	db := SetupDatabase()
	defer db.Close()

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	query := `SELECT post_id, title, moderator_request, admin_response FROM moderator_request`

	// Execute the query within the transaction
	rows, err := tx.Query(query)
	if err != nil {
		tx.Rollback() // Rollback transaction on error
		return nil, fmt.Errorf("error querying moderator requests: %v", err)
	}
	defer rows.Close()

	var requests []ModeratorRequest

	for rows.Next() {
		var request ModeratorRequest

		err := rows.Scan(&request.PostID, &request.Title, &request.ModeratorRequest, &request.AdminResponse)
		if err != nil {
			log.Printf("error scanning row: %v", err)
			tx.Rollback() // Rollback transaction if an error occurs while scanning
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		requests = append(requests, request)
	}

	// Check for errors after iterating through rows
	if err := rows.Err(); err != nil {
		tx.Rollback() // Rollback transaction if there's an iteration error
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback if commit fails
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return requests, nil
}
