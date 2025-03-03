package handlers

import (
	"db"
	"fmt"
	"html/template"
	"middlewares"
	"net/http"
	"net/url"
	"strconv"
)

// IndexHandler handles requests to the root of the server.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	imgErrorMsg := r.URL.Query().Get("error")
	if imgErrorMsg != "" {
		imgErrorMsg, _ = url.QueryUnescape(imgErrorMsg)
	}

	// Different HTTP headers testing
	// w.WriteHeader(http.StatusBadRequest) // 400
	// w.WriteHeader(http.StatusInternalServerError) // 500
	// return

	// Don't forget to get the get notification module on each page.
	var notification []db.Notification
	// Get the user infos
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		session.Username = "Traveler"
		session.Role = "traveler"
		session.UserID = -1
	} else {
		session.Username, _ = db.DecryptData(session.Username)
		notification, _ = db.NotificationsSelect(session.UserID)

	}
	var mostRecentPosts []db.Post
	var mrErr error
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error while getting the post method", http.StatusBadRequest)
			return
		}
		selectedMoods := r.Form["moods"]
		categories := []int{}
		for _, mood := range selectedMoods {
			category, _ := strconv.Atoi(mood)
			categories = append(categories, category)
		}
		mostRecentPosts, mrErr = db.FilterPostsByCategories(categories)
	} else {
		// Get all recent posts.
		mostRecentPosts, mrErr = db.FilterSelectMostRecentPosts()
	}
	if mrErr != nil {
		fmt.Printf("error while getting most recent posts : %v", mrErr)
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
	// Count likes and dislikes.
	mostRecentPosts = countLikesDislikes(mostRecentPosts)
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/index.html",
		"web/templates/tmpl_nav.html",
		"web/templates/tmpl_newpost.html",
		"web/templates/tmpl_categories.html",
		"web/templates/tmpl_categories_selection.html",
		"web/templates/tmpl_lastposts.html",
		"web/templates/tmpl_newcomment.html",
		"web/templates/tmpl_likes_dislikes.html",
		"web/templates/tmpl_likes_dislikes_com.html",
		"web/templates/tmpl_status_posts.html")
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	moods, _ := db.SelectAllCategories()
	indexData := IndexPage{
		Title:             "4mood",
		Nav:               NavData,
		NewPost:           NewPostData,
		Moods:             moods,
		MostRecentPosts:   mostRecentPosts,
		UserID:            session.UserID,
		UserUsername:      session.Username,
		UserRole:          session.Role,
		NotificationCount: countUnReadNotifications(notification),
		CurrentPage:       "index",
		ErrorMsgs:         imgErrorMsg,
	}
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, indexData)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}
