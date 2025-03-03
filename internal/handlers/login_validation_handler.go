package handlers

import (
	"db"
	"html/template"
	"middlewares"
	"net/http"
	"net/url"
	"strings"
)

// Structure to pass data to the template
type BanMessageData struct {
	Username string
	Message  string
}

func LoginValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the method is POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve form data
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")

	// Call the database validation function
	user, err := db.UserSelectLogin(email, password)
	if err != nil {
		// On error, redirect to the login page with an error message
		errorMsg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/login?error="+errorMsg, http.StatusSeeOther)
		return
	}

	// Check if the user is banned
	if user.Role == "banned" {
		// Parse the ban message template
		tmpl, tmplErr := template.ParseFiles("web/pages/login.html", "web/templates/tmpl_message_usr_ban.html","web/templates/tmpl_login.html")
		if tmplErr != nil {
			http.Error(w, "Error rendering the access denied page", http.StatusInternalServerError)
			return
		}

		// Render the defined ban message template
		tmpl.ExecuteTemplate(w, "ban_message", nil)
		return
	}

	// Create a session for the non-banned user
	middlewares.CreateSession(w, user.ID, user.Username, user.Role)

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
