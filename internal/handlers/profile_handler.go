package handlers

import (
	"db"
	"fmt"
	"html/template"
	"log"
	"middlewares"
	"net/http"
)

type ProfileData struct {
	Username          string
	UserRole          string
	UserID            int
	CreatedAt         string
	MostRecentPosts   []db.Post
	Nav               NavTmpl
	NotificationCount int
	Error             string
	Success           string
	ShowUpdateForm    bool
	CurrentPage       string
}

// Fonction handler
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error fetching user from database")
		return
	}
	notifications, _ := db.NotificationsSelect(session.UserID)
	userName, err := db.DecryptData(session.Username)
	if err != nil {
		http.Error(w, "Error on the Decrypt username", http.StatusNotFound)
		log.Println("Error fetching username from database")
		return
	}
	activities, err := db.FilterUserPosts(session.UserID)
	if err != nil {
		fmt.Printf("error while getting user activity : %v", err)
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
	activities = countLikesDislikes(activities)

	showUpdateForm := r.URL.Query().Get("update") == "true"

	data := ProfileData{
		Username:          userName,
		UserRole:          session.Role,
		UserID:            session.UserID,
		Nav:               NavData,
		MostRecentPosts:   activities,
		NotificationCount: countUnReadNotifications(notifications),
		Error:             r.URL.Query().Get("error"),
		Success:           r.URL.Query().Get("success"),
		ShowUpdateForm:    showUpdateForm,
		CurrentPage:       "profile",
	}

	tmpl, err := template.ParseFiles(
		"web/pages/profile.html",
		"web/templates/tmpl_nav.html",
		"web/templates/tmpl_updateProfile.html",
		"web/templates/tmpl_lastposts.html",
		"web/templates/tmpl_newcomment.html",
		"web/templates/tmpl_likes_dislikes.html",
		"web/templates/tmpl_likes_dislikes_com.html",
		"web/templates/tmpl_user_request.html",
		"web/templates/tmpl_status_posts.html",
	)

	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println("Error parsing templates:", err)
		return
	}

	// Rendu du template avec les données
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}
