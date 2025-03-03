package handlers

import (
	"db"
	"html/template"
	"middlewares"
	"net/http"
)

func Err404Handler(w http.ResponseWriter, r *http.Request) {
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
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/404.html",
		"web/templates/tmpl_nav.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	err404Data := IndexPage{
		Title:             "4mood",
		Nav:               NavData,
		UserID:            session.UserID,
		UserUsername:      session.Username,
		UserRole:          session.Role,
		NotificationCount: countUnReadNotifications(notification),
	}
	w.WriteHeader(http.StatusNotFound)
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, err404Data)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}

func Err429Handler(w http.ResponseWriter, r *http.Request) {
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
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/429.html",
		"web/templates/tmpl_nav.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	err404Data := IndexPage{
		Title:             "4mood",
		Nav:               NavData,
		UserID:            session.UserID,
		UserUsername:      session.Username,
		UserRole:          session.Role,
		NotificationCount: countUnReadNotifications(notification),
	}
	w.WriteHeader(http.StatusTooManyRequests)
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, err404Data)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}

func Err400Handler(w http.ResponseWriter, r *http.Request) {
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
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/400.html",
		"web/templates/tmpl_nav.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	err404Data := IndexPage{
		Title:             "4mood",
		Nav:               NavData,
		UserID:            session.UserID,
		UserUsername:      session.Username,
		UserRole:          session.Role,
		NotificationCount: countUnReadNotifications(notification),
	}
	w.WriteHeader(http.StatusBadRequest)
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, err404Data)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}

func Err500Handler(w http.ResponseWriter, r *http.Request) {
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
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/500.html",
		"web/templates/tmpl_nav.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	err404Data := IndexPage{
		Title:             "4mood",
		Nav:               NavData,
		UserID:            session.UserID,
		UserUsername:      session.Username,
		UserRole:          session.Role,
		NotificationCount: countUnReadNotifications(notification),
	}
	w.WriteHeader(http.StatusInternalServerError)
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, err404Data)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}
