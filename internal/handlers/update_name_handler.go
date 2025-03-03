package handlers

import (
	"db"
	"middlewares"
	"net/http"
	"net/url"
)

func UpdateNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	session := middlewares.GetCookie(w, r)
	if session.UserID == 0 {
		http.Error(w, "Utilisateur non connecté", http.StatusUnauthorized)
		return
	}
	newName := r.FormValue("new_name")
	encryptedNewName, err := db.UserUpdateName(session.UserID, newName)
	if err != nil {
		http.Redirect(w, r, "/profile?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	// Mettre à jour la session
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionID := cookie.Value
		middlewares.DeleteSession(sessionID)
	}
	middlewares.CreateSession(w, session.UserID, encryptedNewName, session.Role)
	// Rediriger vers la page de profil avec un message de succès
	http.Redirect(w, r, "/profile?success=true", http.StatusSeeOther)
}
