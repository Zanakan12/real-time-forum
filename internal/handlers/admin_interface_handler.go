package handlers

import (
	"database/sql"
	"db"
	"fmt"
	"html/template"
	"log"
	"middlewares"
	"net/http"
	"strconv"
)

// AdminInterfaceHandler handles the display of the admin interface and manages user updates/deletions.
func AdminInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user session and decrypt the username
	session := middlewares.GetCookie(w, r)
	userName, err := db.DecryptData(session.Username)

	// Ensure that only an Administrator or a specific user ("admin") can access the admin page
	if session.Role != "admin" && userName != "Zanakan12" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err != nil {
		http.Error(w, "Error on decrypting username", http.StatusNotFound)
		log.Println("Error fetching username from database:", err)
		return
	}
	notifications, _ := db.NotificationsSelect(session.UserID)

	// Handle POST requests for deletion or role updates
	if r.Method == "POST" {
		// Handle user deletion
		if deleteID := r.FormValue("delete_id"); deleteID != "" {
			userID, err := strconv.Atoi(deleteID)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				log.Println("Error parsing user ID:", err)
				return
			}

			err = db.DeleteUser(userID)
			if err != nil {
				http.Error(w, "Error deleting user", http.StatusInternalServerError)
				log.Println("Error deleting user:", err)
				return
			}
		}

		// Handle user role update
		if userIDStr, role := r.FormValue("user_id"), r.FormValue("role"); userIDStr != "" && role != "" {
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				log.Println("Error parsing user ID:", err)
				return
			}

			err = db.UserUpdateRole(userID, role)
			if err != nil {
				http.Error(w, "Error updating user role", http.StatusInternalServerError)
				log.Println("Error updating user role:", err)
				return
			} else if role == "banned" {
				db.PostDelete(userID)
			}
			// Mettre à jour la session
			cookie, err := r.Cookie("session_id")
			if err == nil {
				sessionID := cookie.Value
				session, _ := middlewares.GetSession(sessionID)
				if session.UserID == userID {
					middlewares.DeleteSession(sessionID)
					middlewares.CreateSession(w, session.UserID, session.Username, role)
				}
			}
		}

		if deleteMood := r.FormValue("moodID"); deleteMood != "" {
			userID, err := strconv.Atoi(deleteMood)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				log.Println("Error parsing user ID:", err)
				return
			}

			err = db.DeleteCategory(userID)
			if err != nil {
				http.Error(w, "Error deleting user", http.StatusInternalServerError)
				log.Println("Error deleting user:", err)
				return
			}
		}

		if addMood := r.FormValue("emoji"); addMood != "" {
			err := db.AddCategory(addMood)
			if err != nil {
				http.Error(w, "Error adding mood", http.StatusInternalServerError)
				log.Println("Error adding mood:", err)
				return
			}
		}

		if postID := r.FormValue("post_id"); postID != "" {
			status := r.FormValue("status")
			fmt.Printf("Updating status for post %s to %s\n", postID, status)
			id, _ := strconv.Atoi(postID)
			err := db.UpdatePostStatus(id, status) // Assuming UpdatePostStatus accepts postID and status
			if err != nil {
				http.Error(w, "Error updating post status", http.StatusInternalServerError)
				log.Println("Error updating post status:", err)
				return
			}
			err = db.RequestToAdmin(id,"", "", sql.NullString{String: status, Valid: true})
			if err != nil {
				log.Printf("Error updating post status to mod table: %v", err)
				http.Error(w, "Failed to update post status to mod table", http.StatusInternalServerError)
				return
			}
		}

		if postID := r.FormValue("deletepost_id"); postID != "" {
			id, _ := strconv.Atoi(postID)
			err := db.PostDelete(id)
			if err != nil {
				http.Error(w, "Error deleting post", http.StatusInternalServerError)
				log.Println("Error deleting post:", err)
				return
			}
			err = db.RequestToAdmin(id,"", "", sql.NullString{String: "delete", Valid: true})
			if err != nil {
				log.Printf("Error updating post status to mod table: %v", err)
				http.Error(w, "Failed to update post status to mod table", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	// Load and display the user management view for GET requests
	users, err := db.UserSelect(nil)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		log.Println("Error fetching users from database:", err)
		return
	}
	// Decrypt usernames and prepare data
	for i, user := range users {
		decryptedUsername, err := db.DecryptData(user.Username)
		decryptedUserEmail, err := db.DecryptData(user.Email)
		if err != nil {
			log.Println("Error decrypting username, skipping:", err)
			continue
		}
		users[i].Username = decryptedUsername
		users[i].Email = decryptedUserEmail
	}
	mood, err := db.SelectAllCategories()
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		log.Println("Error fetching category from database:", err)
	}

	dbSignaledPosts, err := db.DisplaySignaledStatus()
	if err != nil {
		log.Println("Error fetching signaled posts:", err)
		http.Error(w, "Error fetching signaled posts", http.StatusInternalServerError)
		return
	}
	// for Put status of post on admin interface
	var signaledPosts []db.Post
	for _, dbPost := range dbSignaledPosts {
		// Decrypt the title before appending
		decryptedTitle, err := db.DecryptData(dbPost.Title)
		if err != nil {
			log.Println("Error decrypting title:", err)
			continue // Skip this post if there's an error in decryption
		}

		// Append the post with the decrypted title
		signaledPosts = append(signaledPosts, db.Post{
			ID:     dbPost.ID,
			Title:  decryptedTitle,
			Status: dbPost.Status,
		})
	}
	data := struct {
		Username          string
		UserRole          string
		Nav               interface{}
		Users             []db.User
		Moods             []db.Category
		NotificationCount int
		Posts             []db.Post // Add this field for posts
		CurrentPage       string
	}{
		Username:          userName,
		UserRole:          session.Role,
		Nav:               NavData,
		Users:             users,
		Moods:             mood,
		NotificationCount: countUnReadNotifications(notifications),
		Posts:             signaledPosts, // Pass the posts to the template
		CurrentPage:       "admin",
	}
	// Load and execute the template
	tmpl, err := template.ParseFiles(
		"web/templates/tmpl_nav.html",
		"web/pages/admin.html",
		"web/templates/tmpl_manage_users.html",
		"web/templates/tmpl_categories_manage.html",
		"web/templates/tmpl_status_posts.html",
		"web/templates/tmpl_display_signaled_posts.html",
	)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println("Error parsing templates:", err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "admin.html", data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
	}
}
