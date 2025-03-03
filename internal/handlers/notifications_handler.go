package handlers

import (
	"db"
	"fmt"
	"html/template"
	"log"
	"middlewares"
	"net/http"
)

type NotificationsData struct {
	Title             string
	UserRole          string
	UserID            int
	CreatedAt         string
	Comments          []db.Comment
	LikesDislikes     []db.LikesDislikes
	Nav               NavTmpl
	NotificationCount int
	CurrentPage		  string
}

// Fonction handler
func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Println("Error fetching user from database")
		return
	}
	notifications, err := db.NotificationsSelect(session.UserID)
	if err != nil {
		fmt.Printf("error while getting user notifications : %v", err)
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
	var comments []db.Comment
	var likesDislikes []db.LikesDislikes
	for _, notification := range notifications {
		if notification.CommentID.Valid {
			comment, _ := db.CommentSelectByID(notification.CommentID.Int64)
			comment.PostTitle, _ = db.PostTitleSelectById(comment.PostID)
			comments = append(comments, comment)
		} else if notification.LikeDislikeID.Valid {
			likeDislike, _ := db.LikesSelectByID(notification.LikeDislikeID.Int64)
			likeDislike.PostTitle, _ = db.PostTitleSelectById(int(likeDislike.PostID.Int64))
			user, _ := db.UserSelectById(likeDislike.UserID)
			likeDislike.Username = user.Username
			likesDislikes = append(likesDislikes, likeDislike)
		}
		err := db.NotificationsUpdateIsRead(int(notification.ID))
		if err != nil {
			fmt.Println("error while updating notifications: %v", err)
		}
	}
	//comments = countLikesDislikesComments(comments)
	data := NotificationsData{
		Title:             "Notifications",
		UserRole:          session.Role,
		UserID:            session.UserID,
		Nav:               NavData,
		Comments:          comments,
		LikesDislikes:     likesDislikes,
		NotificationCount: countUnReadNotifications(notifications),
		CurrentPage:       "notifications",
	}

	tmpl, err := template.ParseFiles(
		"web/pages/notifications.html",
		"web/templates/tmpl_nav.html",
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

func countUnReadNotifications(notifications []db.Notification) int {
	unReadCount := 0
	for _, notification := range notifications {
		if !notification.IsRead {
			unReadCount++
		}
	}
	return unReadCount
}
