package handlers

import (
	"middlewares"
	"net/http"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	sessionID, exists := middlewares.SessionExists(session.UserID)
	if exists {
		middlewares.DeleteSession(sessionID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
