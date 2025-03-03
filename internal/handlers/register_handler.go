package handlers

import (
	"html/template"
	"net/http"
	"net/url"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer le message d'erreur s'il existe
	errorMsg := r.URL.Query().Get("error")
	if errorMsg != "" {
		errorMsg, _ = url.QueryUnescape(errorMsg)
	}
	// Parse the templates from files.
	tmpl, err := template.ParseFiles(
		"web/pages/register.html",
		"web/templates/tmpl_register.html",
		"web/templates/tmpl_nav.html",
	)

	if err != nil {
		http.Error(w, "Internal Server Error (Error parsing templates)", http.StatusInternalServerError)
		return
	}
	registerData := RegisterPage{
		Nav:               NavData,
		Title:             "Forum | Register",
		NotificationCount: 0,
		UserRole:          "traveler",
		Register:          RegisterData,
		CurrentPage:       "register",
	}
	registerData.Register.Error = errorMsg
	// Execute the template and pass any data needed (nil here for simplicity).
	err = tmpl.Execute(w, registerData)
	if err != nil {
		http.Error(w, "Internal Server Error (Error executing template)", http.StatusInternalServerError)
		return
	}
}
