package db

import (
	"database/sql"
	"fmt"
)

type category struct {
	Name string
}

var categories = []category{
	{Name: "😊"},
	{Name: "😢"},
	{Name: "😡"},
	{Name: "🤩"},
}

func createCategoriesTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
       	name TEXT
    );`

	executeSQL(db, createTableSQL)
}

func createPostCategoriesTable(db *sql.DB) {
	// SQL statement to create the table that links posts and categories
	createTableSQL := `CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER NOT NULL ,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id)
);`
	executeSQL(db, createTableSQL)
}

func CategoryInsertDefault() error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	for i := range categories {
		insertSQL := `INSERT INTO categories (name) VALUES (?)`
		_, err = tx.Exec(insertSQL, categories[i].Name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing statement: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func categorySelectByPostId(postID int, db *sql.DB) ([]Category, error) {

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	query := `
        SELECT c.id, c.name
        FROM post_categories pc
        JOIN categories c ON pc.category_id = c.id
        WHERE pc.post_id = ?
        `

	rows, err := tx.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("error executing category query: %v", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("error scanning category row: %v", err)
		}
		categories = append(categories, category)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return categories, nil
}

func categoriesInsertByPostId(postID int64, categories []int, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("error rolling back transaction: %v (original error: %v)", rbErr, err)
			}
		}
	}()
	stmt, err := tx.Prepare(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	for _, category := range categories {
		_, err = stmt.Exec(postID, category)
		if err != nil {
			return fmt.Errorf("error executing statement: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func DeleteCategory(id_category int) error {
	db := SetupDatabase()
	defer db.Close()

	// Démarrer la transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de la transaction : %v", err)
	}
	// Préparer la requête DELETE avec une condition WHERE
	deleteSQL := `DELETE FROM categories WHERE id = ?`

	// Exécuter la requête en passant le nom de la catégorie à supprimer
	_, err = tx.Exec(deleteSQL, id_category)
	if err != nil {
		tx.Rollback() // Annuler la transaction en cas d'erreur
		return fmt.Errorf("erreur lors de l'exécution de la requête : %v", err)
	}

	// Valider la transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("erreur lors de la validation de la transaction : %v", err)
	}

	return nil
}

func AddCategory(newCategory string) error {
	db := SetupDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	insertSQL := `INSERT INTO categories (name) VALUES (?)`
	_, err = tx.Exec(insertSQL, newCategory)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing statement: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func SelectAllCategories() ([]Category, error) {
	db := SetupDatabase()
	defer db.Close()

	// Prepare the SELECT query to get all categories
	query := `SELECT id, name FROM categories`

	// Execute the query and get the rows
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Prepare a slice to store the categories
	var categories []Category

	// Loop through the rows and scan each row into a Category struct
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		// Append the category to the slice
		categories = append(categories, category)
	}

	// Check for any error during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Return the slice of categories
	return categories, nil
}
