package handlers

import (
	"db"
	"middlewares" // Assurez-vous d'importer votre package middlewares
	"net/http"
	"strconv"
)

func CommentUpdateValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode est POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	commentID := r.FormValue("comment_id")
	content := r.FormValue("content")

	// Convertir l'ID du commentaire en entier
	convCommentID, err := strconv.Atoi(commentID)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Récupérer la session de l'utilisateur
	session := middlewares.GetCookie(w, r)
	if session.UserID == 0 {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupérer les informations du commentaire
	comment, err := db.CommentSelectByID(int64(convCommentID))
	if err != nil {
		http.Error(w, "Error retrieving comment information", http.StatusInternalServerError)
		return
	}

	// Vérifier si l'utilisateur est le propriétaire du commentaire
	if comment.UserID != session.UserID {
		http.Error(w, "Non autorisé à modifier ce commentaire", http.StatusForbidden)
		return
	}

	// Appeler la fonction de mise à jour du commentaire
	err = db.CommentUpdate(convCommentID, session.UserID, comment.PostID, content)
	if err != nil {
		http.Error(w, "Error updating comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
