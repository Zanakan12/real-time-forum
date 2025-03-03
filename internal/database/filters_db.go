package db

import (
	"fmt"
	"strings"
)

// GetAllRecentPosts retrieves all posts from the database using a transaction
func FilterSelectMostRecentPosts() ([]Post, error) {
	db := SetupDatabase()
	defer db.Close()
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	// Ensure rollback in case of an error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rethrow panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	query := `
        SELECT p.id, p.user_id, p.title, p.body, p.status, p.created_at, p.updated_at,
        u.id, u.email, u.username, u.role, u.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC`


		rows, err := tx.Query(query)
		if err != nil {
			return nil, fmt.Errorf("error executing query: %v", err)
		}
		defer rows.Close()
		var posts []Post
		for rows.Next() {
			post := Post{}
			user := User{}
			if err := rows.Scan(
				&post.ID, &post.UserID, &post.Title, &post.Body, &post.Status, &post.CreatedAt, &post.UpdatedAt,
				&user.ID, &user.Email, &user.Username, &user.Role, &user.CreatedAt,
			); err != nil {
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
			post.Title, _ = DecryptData(post.Title)
			post.Body, _ = DecryptData(post.Body)
			post.User = user
			post.User.Username, _ = DecryptData(post.User.Username)
			posts = append(posts, post)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %v", err)
		}
		for i, post := range posts {
			posts[i].Categories, err = categorySelectByPostId(post.ID, db)
			if err != nil {
				return nil, fmt.Errorf("error parsing categories: %v", err)
			}
			posts[i].Comments, err = CommentSelectByPostID(post.ID, db)
			if err != nil {
				return nil, fmt.Errorf("error parsing comments: %v", err)
			}
			posts[i].LikesDislikes, err = LikesSelectByPostID(post.ID, db)
			if err != nil {
				return nil, fmt.Errorf("error parsing likes dislikes: %v", err)
			}
			image, err := ImageSelectByPostID(post.ID, db)
			if err != nil {
				return nil, fmt.Errorf("error parsing images: %v", err)
			}
			if len(image) != 0 {
				decryptedImagePath, _ := DecryptData(image[0].FilePath)
				posts[i].ImagePath = decryptedImagePath
			}
	
		}
	
		// Commit the transaction
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("error committing transaction: %v", err)
		}
		return posts, nil
	}

func FilterSelectMostLikedPosts() ([]PostTendance, error) {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("erreur lors du démarrage de la transaction : %v", err)
	}

	countSQL := `
        SELECT ld.post_id, COUNT(ld.id) as count
        FROM likes_dislikes ld
        GROUP BY ld.post_id
        ORDER BY count DESC
    `

	rows, err := tx.Query(countSQL)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erreur lors de l'exécution de la requête : %v", err)
	}
	defer rows.Close()

	var tendances []PostTendance
	for rows.Next() {
		var t PostTendance
		if err := rows.Scan(&t.PostID, &t.Count); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("erreur lors de la lecture des résultats : %v", err)
		}
		tendances = append(tendances, t)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erreur après la lecture des lignes : %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("erreur lors de la validation de la transaction : %v", err)
	}

	return tendances, nil
}

/*
func FilterSelectActivityByUserID(userID int) ([]Activity, error) {
	db := SetupDatabase()
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
		return nil, fmt.Errorf("error querying posts: %v", err)
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
		return nil, fmt.Errorf("error querying comments: %v", err)
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
		return nil, fmt.Errorf("error querying likes/dislikes: %v", err)
	}
	defer likeRows.Close()

	// Helper function to scan rows
	scanRows := func(rows *sql.Rows) error {
		for rows.Next() {
			var a Activity
			if err := rows.Scan(&a.Type, &a.Content, &a.Timestamp); err != nil {
				return err
			}
			decryptedContent, err := DecryptData(a.Content)
			if err != nil {
				return fmt.Errorf("unable to decrypt content: %v", err)
			}
			a.Content = decryptedContent
			activities = append(activities, a)
		}
		return rows.Err()
	}

	// Scan all result sets
	if err := scanRows(postRows); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error scanning post rows: %v", err)
	}
	if err := scanRows(commentRows); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error scanning comment rows: %v", err)
	}
	if err := scanRows(likeRows); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error scanning like/dislike rows: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return activities, nil
}
*/

func FilterUserPosts(userID int) ([]Post, error) {
	db := SetupDatabase()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	// Ensure rollback in case of an error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rethrow panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Query to get posts created, commented on, or liked by the user
	query := `
    SELECT DISTINCT p.id, p.user_id, p.title, p.body, p.status, p.created_at, p.updated_at,
    u.id, u.email, u.username, u.role, u.created_at
    FROM posts p
    JOIN users u ON p.user_id = u.id
    LEFT JOIN comments c ON p.id = c.post_id
    LEFT JOIN likes_dislikes ld ON p.id = ld.post_id
    WHERE p.user_id = ? OR c.user_id = ? OR ld.user_id = ?
    ORDER BY p.created_at DESC`

	rows, err := tx.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		post := Post{}
		user := User{}
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Body, &post.Status, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Email, &user.Username, &user.Role, &user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		post.Title, _ = DecryptData(post.Title)
		post.Body, _ = DecryptData(post.Body)
		post.User = user
		post.User.Username, _ = DecryptData(post.User.Username)
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	for i, post := range posts {
		posts[i].Categories, err = categorySelectByPostId(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing categories: %v", err)
		}
		posts[i].Comments, err = CommentSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing comments: %v", err)
		}
		posts[i].LikesDislikes, err = LikesSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing likes dislikes: %v", err)
		}
		image, err := ImageSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing images: %v", err)
		}
		if len(image) != 0 {
			decryptedImagePath, _ := DecryptData(image[0].FilePath)
			posts[i].ImagePath = decryptedImagePath
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return posts, nil
}

func FilterPostsByCategories(categoryIDs []int) ([]Post, error) {
	db := SetupDatabase()
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	// Ensure rollback in case of an error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rethrow panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the query
	query := `
    SELECT DISTINCT p.id, p.user_id, p.title, p.body, p.status, p.created_at, p.updated_at,
    u.id, u.email, u.username, u.role, u.created_at
    FROM posts p
    JOIN users u ON p.user_id = u.id
    JOIN post_categories pc ON p.id = pc.post_id
    WHERE pc.category_id IN (`

	// Create a slice to hold all query arguments
	args := make([]interface{}, 0)

	// Add placeholders for category IDs
	placeholders := make([]string, len(categoryIDs))
	for i, id := range categoryIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query += strings.Join(placeholders, ",") + ")"

	query += " ORDER BY p.created_at DESC"

	// Execute the query
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		post := Post{}
		user := User{}
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Body, &post.Status, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Email, &user.Username, &user.Role, &user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		post.Title, _ = DecryptData(post.Title)
		post.Body, _ = DecryptData(post.Body)
		post.User = user
		post.User.Username, _ = DecryptData(post.User.Username)
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	for i, post := range posts {
		posts[i].Categories, err = categorySelectByPostId(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing categories: %v", err)
		}
		posts[i].Comments, err = CommentSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing comments: %v", err)
		}
		posts[i].LikesDislikes, err = LikesSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing likes dislikes: %v", err)
		}
		image, err := ImageSelectByPostID(post.ID, db)
		if err != nil {
			return nil, fmt.Errorf("error parsing images: %v", err)
		}
		if len(image) != 0 {
			decryptedImagePath, _ := DecryptData(image[0].FilePath)
			posts[i].ImagePath = decryptedImagePath
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return posts, nil
}
