package handlers

import (
	"encoding/json"
	"net/http"
	"middlewares"
	"db"
)

// GetUserHandler retourne l'utilisateur connecté
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	if session.Username == "traveler" {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	userName, err := db.DecryptData(session.Username)
	if err != nil || userName == "" {
		http.Error(w, "Session invalide", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"username": userName})
}
