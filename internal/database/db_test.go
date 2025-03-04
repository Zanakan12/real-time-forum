package db

import (
	"config"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestDatabaseAuthentication(t *testing.T) {
	dbPath := "test_forum.db"

	// Test with correct credentials
	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	dbCorrect, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Failed to open database with correct credentials: %v", err)
	}
	defer dbCorrect.Close()

	// Verify connection with correct credentials
	err = dbCorrect.Ping()
	if err != nil {
		t.Errorf("Connection failed with correct credentials: %v", err)
	} else {
		t.Log("Connection successful with correct credentials")
	}

	// Test with incorrect credentials
	connStringIncorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=wrongpassword&_auth_crypt=sha256", dbPath)
	dbIncorrect, err := sql.Open("sqlite3", connStringIncorrect)
	if err != nil {
		t.Fatalf("Failed to open database with incorrect credentials: %v", err)
	}
	defer dbIncorrect.Close()

	// Verify connection fails with incorrect credentials
	err = dbIncorrect.Ping()
	if err == nil {
		t.Error("Connection unexpectedly succeeded with incorrect credentials")
	} else {
		t.Log("Connection correctly failed with incorrect credentials:", err)
	}
}

// RegisterUser registers a new user in the database.
func TestRegisterUser(t *testing.T) {
	// Initialize data
	email := "myemail@richman.fr"
	username := "richguy"
	password := "12345678910"
	role := "user"

	dbPath := "test_forum.db"

	// Test with correct credentials
	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	db, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createUsersTable(db)

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("error starting transaction: %v", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("error hashing password: %v", err)
	}

	/******************************************************************
		* When you'll need to compare the input password of the user while
		* login you'll youse another function with bcrypt which is like :
	*******************************************************************/

	/*
		func verifyPassword(hashedPassword, password string) bool {
			err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
			return err == nil
		}
	*/

	// DATA ENCRYPTION:
	encryptedEmail, err := EncryptData(email)
	if err != nil {
		t.Errorf("error encrypting email: %v", err)
	}
	encryptedUsername, err := EncryptData(username)
	if err != nil {
		t.Errorf("error encrypting username: %v", err)
	}

	insertSQL := `INSERT INTO users (email, username, password, role) VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(insertSQL, encryptedEmail, encryptedUsername, hashedPassword, role)
	if err != nil {
		tx.Rollback()
		t.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		t.Errorf("error committing transaction: %v", err)
	}
}

func TestEncryptedImageData(t *testing.T) {
	dbPath := "test_forum.db"

	// Test with correct credentials
	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	db, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create images table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			file_path TEXT,
			file_size INTEGER,
			created_at TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Error creating images table: %v", err)
	}

	// Test data
	postID := 1
	filePath := "/path/to/image.jpg"
	fileSize := 1024
	createdAt := time.Now().UTC()

	// Encrypt file path
	encryptedFilePath, err := EncryptData(filePath)
	if err != nil {
		t.Fatalf("Error encrypting file path: %v", err)
	}
	t.Logf("Encrypted File Path %s: ", encryptedFilePath)

	// Insert encrypted data
	_, err = db.Exec(`
		INSERT INTO images (post_id, file_path, file_size, created_at)
		VALUES (?, ?, ?, ?)
	`, postID, encryptedFilePath, fileSize, createdAt)
	if err != nil {
		t.Fatalf("Error inserting encrypted data: %v", err)
	}

	// Retrieve and decrypt data
	var retrievedID int
	var retrievedPostID int
	var retrievedEncryptedFilePath string
	var retrievedFileSize int
	var retrievedCreatedAt time.Time

	err = db.QueryRow(`
		SELECT id, post_id, file_path, file_size, created_at
		FROM images
		WHERE post_id = ?
	`, postID).Scan(&retrievedID, &retrievedPostID, &retrievedEncryptedFilePath, &retrievedFileSize, &retrievedCreatedAt)
	if err != nil {
		t.Fatalf("Error retrieving data: %v", err)
	}

	// Decrypt file path
	decryptedFilePath, err := DecryptData(retrievedEncryptedFilePath)
	if err != nil {
		t.Fatalf("Error decrypting file path: %v", err)
	}

	// Verify decrypted data
	if decryptedFilePath != filePath {
		t.Errorf("Decrypted file path does not match original. Got %s, want %s", decryptedFilePath, filePath)
	}
	if retrievedPostID != postID {
		t.Errorf("Retrieved post ID does not match. Got %d, want %d", retrievedPostID, postID)
	}
	if retrievedFileSize != fileSize {
		t.Errorf("Retrieved file size does not match. Got %d, want %d", retrievedFileSize, fileSize)
	}
	t.Log("Encryption and decryption of image data successful")
	t.Logf("Decrypted File Path: %s", decryptedFilePath)
}

func TestInsertImage(t *testing.T) {
	//Intialize Data
	filePath := "path/to/img.jpg" // change ext to test wrong ext.
	postId := 4
	fileSize := 2456
	dbPath := "test_forum.db"

	// Verify if the file has an extension.
	fileExt := strings.ToLower(filepath.Ext(filePath))
	if fileExt == "" {
		t.Errorf("the file doesn't have an extension")
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
		t.Errorf("the %s extension is not valid", fileExt)
	}

	// Test with correct credentials
	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	db, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createImagesTable(db)

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("error starting transaction: %v", err)
	}

	encryptedFilePath, err := EncryptData(filePath)
	if err != nil {
		t.Errorf("error encrypting title: %v", err)
	}

	createSQL := `INSERT INTO images (post_id, file_path, file_size) VALUES (?, ?, ?)`
	_, err = tx.Exec(createSQL, postId, encryptedFilePath, fileSize)

	if err != nil {
		tx.Rollback()
		t.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		t.Errorf("error committing transaction: %v", err)
	}
}

// RegisterUser registers a new user in the database.
func TestNotifications(t *testing.T) {
	userID := 1
	dbPath := "test_forum.db"

	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	db, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createUsersTable(db)
	createPostsTable(db)
	createCommentsTable(db)
	createCategoriesTable(db)
	createPostCategoriesTable(db)
	createLikesDislikesTable(db)
	createImagesTable(db)
	createNotificationsTable(db)

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("error starting transaction: %v", err)
	}

	// create a post:
	createSQL := `INSERT INTO posts (user_id, title, body, status) VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(createSQL, userID, "title", "content", "published")

	if err != nil {
		tx.Rollback()
		t.Errorf("error executing statement: %v", err)
	}

	// create alike on a post :
	// Like/dislike on a post
	_, err = tx.Exec("INSERT INTO likes_dislikes (user_id, post_id, is_like) VALUES (?, ?, ?)", 2, 1, true)
	if err != nil {
		tx.Rollback()
		t.Errorf("error inserting like/dislike for post: %v", err)
	}

	// create a comment:
	// Insert comment
	insertSQL := `INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)`
	result, err := tx.Exec(insertSQL, 1, 2, "haha well its good")
	if err != nil {
		tx.Rollback()
		t.Errorf("error executing statement: %v", err)
	}

	// Like/dislike on a comment:
	_, err = tx.Exec("INSERT INTO likes_dislikes (user_id, comment_id, is_like) VALUES (?, ?, ?)", 2, 1, false)
	if err != nil {
		tx.Rollback()
		t.Errorf("error inserting like/dislike for post: %v", err)
	}

	// Get the ID of the inserted comment
	commentID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		t.Errorf("error getting last insert ID: %v", err)
	}

	// Get the user ID of the post owner
	var postOwnerID int64
	err = tx.QueryRow("SELECT user_id FROM posts WHERE id = ?", 1).Scan(&postOwnerID)
	if err != nil {
		tx.Rollback()
		t.Errorf("error getting post owner ID: %v", err)
	}

	// Insert notification for the post owner
	_, err = tx.Exec("INSERT INTO notifications (user_id, comment_id) VALUES (?, ?)", postOwnerID, commentID)
	if err != nil {
		tx.Rollback()
		t.Errorf("error inserting notification: %v", err)
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
		t.Errorf("error querying notifications: %v", err)
	}
	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.CommentID, &n.LikeDislikeID, &n.IsRead, &n.CreatedAt)
		if err != nil {
			tx.Rollback()
			t.Errorf("error scanning notification row: %v", err)
		}
		notifications = append(notifications, n)
	}

	// testing activity :
	userID = 2
	var activities []Activity

	// Query for posts
	postRows, err := tx.Query(`
        SELECT 'post' as type, title as content, created_at 
        FROM posts 
        WHERE user_id = ?
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		tx.Rollback()
		t.Errorf("error querying posts: %v", err)
	}
	defer postRows.Close()

	// Query for comments
	commentRows, err := tx.Query(`
        SELECT 'comment' as type, content, created_at 
        FROM comments 
        WHERE user_id = ?
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		tx.Rollback()
		t.Errorf("error querying comments: %v", err)
	}
	defer commentRows.Close()

	// Query for likes/dislikes
	likeRows, err := tx.Query(`
        SELECT 
            CASE WHEN is_like THEN 'like' ELSE 'dislike' END as type,
            CASE 
                WHEN post_id IS NOT NULL THEN 'post'
                WHEN comment_id IS NOT NULL THEN 'comment'
            END as content,
            created_at
        FROM likes_dislikes
        WHERE user_id = ?
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		tx.Rollback()
		t.Errorf("error querying likes/dislikes: %v", err)
	}
	defer likeRows.Close()

	// Helper function to scan rows
	scanRows := func(rows *sql.Rows) error {
		for rows.Next() {
			var a Activity
			if err := rows.Scan(&a.Type, &a.Content, &a.Timestamp); err != nil {
				return err
			}
			activities = append(activities, a)
		}
		return rows.Err()
	}

	// testing likesdislikes from posts :
	query2 := `SELECT *
              FROM likes_dislikes
              WHERE post_id=?`

	rows2, err := tx.Query(query2, 1)
	if err != nil {
		t.Errorf("error executing query: %v", err)
	}
	defer rows2.Close()

	var likesDislikes []LikesDislikes
	for rows2.Next() {
		var likeDislike LikesDislikes
		if err := rows2.Scan(&likeDislike.ID, &likeDislike.UserID, &likeDislike.PostID, &likeDislike.CommentID, &likeDislike.IsLike, &likeDislike.CreatedAt); err != nil {
			t.Errorf("error scanning row: %v", err)
		}
		likesDislikes = append(likesDislikes, likeDislike)
	}

	query4 := `SELECT *
              FROM likes_dislikes
              WHERE comment_id=?`

	rows10, err := tx.Query(query4, commentID)
	if err != nil {
		t.Errorf("error executing query: %v", err)
	}
	defer rows10.Close()

	var likesDislikes2 []LikesDislikes
	for rows10.Next() {
		var likeDislike LikesDislikes
		if err := rows10.Scan(&likeDislike.ID, &likeDislike.UserID, &likeDislike.PostID, &likeDislike.CommentID, &likeDislike.IsLike, &likeDislike.CreatedAt); err != nil {
			t.Errorf("error scanning row: %v", err)
		}
		likesDislikes2 = append(likesDislikes2, likeDislike)
	}

	// Scan all result sets
	if err := scanRows(postRows); err != nil {
		tx.Rollback()
		t.Errorf("error scanning post rows: %v", err)
	}
	if err := scanRows(commentRows); err != nil {
		tx.Rollback()
		t.Errorf("error scanning comment rows: %v", err)
	}
	if err := scanRows(likeRows); err != nil {
		tx.Rollback()
		t.Errorf("error scanning like/dislike rows: %v", err)
	}

	if err = tx.Commit(); err != nil {
		t.Errorf("error committing transaction: %v", err)
	}
	t.Log(notifications)
	t.Log(activities)
	t.Log(likesDislikes)
	t.Log(likesDislikes2)
}

func TestInsertLikesDislikes(t *testing.T) {
	// Initialize data
	tests := []struct {
		userID    int
		postID    int
		commentID int
		isLike    bool
	}{
		{
			userID:    1,
			postID:    -1,
			commentID: 45,
			isLike:    true,
		},
		{
			userID:    56,
			postID:    5,
			commentID: -1,
			isLike:    false,
		},
	}

	dbPath := "test_forum.db"
	// Test with correct credentials
	connStringCorrect := fmt.Sprintf("%s?_auth&_auth_user=admin&_auth_pass=adminpassword&_auth_crypt=sha256", dbPath)
	db, err := sql.Open("sqlite3", connStringCorrect)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createLikesDislikesTable(db)

	for _, tt := range tests {
		t.Run(fmt.Sprintf("UserID=%d,PostID=%d,CommentID=%d,IsLike=%v", tt.userID, tt.postID, tt.commentID, tt.isLike), func(t *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("Error starting transaction: %v", err)
			}

			var insertSQL string
			var args []interface{}

			if tt.postID == -1 {
				insertSQL = `INSERT INTO likes_dislikes (user_id, comment_id, is_like) VALUES (?, ?, ?)`
				args = []interface{}{tt.userID, tt.commentID, tt.isLike}
			} else if tt.commentID == -1 {
				insertSQL = `INSERT INTO likes_dislikes (user_id, post_id, is_like) VALUES (?, ?, ?)`
				args = []interface{}{tt.userID, tt.postID, tt.isLike}
			} else {
				t.Fatalf("Invalid test case: both postID and commentID are set or both are -1")
			}

			_, err = tx.Exec(insertSQL, args...)
			if err != nil {
				tx.Rollback()
				t.Fatalf("Error executing statement: %v", err)
			}

			if err = tx.Commit(); err != nil {
				t.Fatalf("Error committing transaction: %v", err)
			}

			// Optionally, you can add assertions here to verify the insertion
			// For example, query the database to check if the record was inserted correctly
		})
	}
}
