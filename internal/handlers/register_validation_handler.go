package handlers

import (
	"db"
	"net/http"
	"net/url"
	"strings"
)

func RegisterValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	email, username, password := strings.ToLower(r.FormValue("email")), r.FormValue("username"), r.FormValue("password")
	err := db.UserInsertRegister(email, username, password, "user")
	if err != nil {
		// En cas d'erreur, rediriger vers la page de connexion
		errorMsg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/register?error="+errorMsg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
