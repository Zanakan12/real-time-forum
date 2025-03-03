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

func ModeratorPowerHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	// Ensure that only an Administrator or a specific user ("admin") can access the admin page
	if session.Role != "moderator" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Extract the post ID from the form or query parameter
	idStr := r.FormValue("post_id")
	status := r.FormValue("status")
	title := r.FormValue("title")
	// Convert the post ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Call the function to update the status
	err = db.UpdatePostStatus(id, status)
	if err != nil {
		log.Printf("Error updating post status: %v", err)
		http.Error(w, "Failed to update post status", http.StatusInternalServerError)
		return
	}
	err = db.RequestToAdmin(id, title, status, sql.NullString{String: "", Valid:false})
	if err != nil {
		log.Printf("Error updating post status to mod table: %v", err)
		http.Error(w, "Failed to update post status to mod table", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ModeratorInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch moderator requests from your database (this is a dummy example)
	adminResponse, err := db.DisplayAdminResponse()
	if err != nil {
		log.Printf("Error retrieving moderator requests: %v", err)
		http.Error(w, "Error retrieving moderator requests", http.StatusInternalServerError)
		return
	}
	session := middlewares.GetCookie(w, r)
	userRole := session.Role
	notifications, _ := db.NotificationsSelect(session.UserID)
	// Structure for passing the data to the template
	data := struct {
		Nav               NavTmpl
		ModeratorRequests []db.ModeratorRequest
		UserRole          string
		NotificationCount int
		CurrentPage       string
	}{
		Nav:               NavData,
		ModeratorRequests: adminResponse,
		UserRole:          userRole,
		NotificationCount: countUnReadNotifications(notifications),
		CurrentPage:       "moderator",
	}
	fmt.Println(data.CurrentPage)
	// Parse and load the template files
	tmpl, err := template.ParseFiles(
		"web/templates/tmpl_nav.html",
		"web/pages/moderator_interface.html",
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Execute the template and pass the data
	err = tmpl.ExecuteTemplate(w, "moderator_interface.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}
