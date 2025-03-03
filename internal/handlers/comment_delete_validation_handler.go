package handlers

import (
	"db"
	"net/http"
	"strconv"
)

func CommentDeleteValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode est POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer les données du formulaire
	commentID := r.FormValue("comment_id")
	convCommentID, _ := strconv.Atoi(commentID)

	// Appeler la fonction de validation dans la base de données
	err := db.CommentDelete((convCommentID))
	if err != nil {
		// En cas d'erreur, rediriger vers la page de connexion
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
