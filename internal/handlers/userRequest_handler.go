package handlers

import (
	"db"
	"fmt"
	"middlewares"
	"net/http"
)

func UserValidationRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Assurer que la méthode est bien POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer la session pour obtenir l'ID et le nom d'utilisateur
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Insérer la requête dans la base de données avec l'ID et le nom d'utilisateur
	err := db.RequestInsert(session.UserID, session.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting request: %v", err), http.StatusInternalServerError)
		return
	}

	// Rediriger l'utilisateur après le succès de l'insertion
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
