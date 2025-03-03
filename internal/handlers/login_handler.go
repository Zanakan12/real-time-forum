package handlers

import (
	"html/template"
	"net/http"
	"net/url"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer le message d'erreur s'il existe
	errorMsg := r.URL.Query().Get("error")
	if errorMsg != "" {
		errorMsg, _ = url.QueryUnescape(errorMsg)
	}
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/login.html",
		"web/templates/tmpl_login.html",
		"web/templates/tmpl_nav.html",
		"web/templates/tmpl_message_usr_ban.html",
	)
	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}

	// Assuming this is an empty login form
	loginData := LoginPage{
		Nav:               NavData,
		Title:             "Forum | Login",
		NotificationCount: 0,
		UserRole:          "traveler",
		Login:             LoginData,
		CurrentPage: 	   "login",
	}
	loginData.Login.Error = errorMsg
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, loginData)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}
