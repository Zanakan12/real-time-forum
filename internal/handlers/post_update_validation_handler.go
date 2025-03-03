package handlers

import (
	"db"
	"net/http"
	"strconv"
)

func PostUpdateValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode est POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer les données du formulaire
	postID := r.FormValue("post_id")
	content := r.FormValue("content")
	convPostID, _ := strconv.Atoi(postID)

	// Appeler la fonction de validation dans la base de données
	err := db.PostUpdateContent(convPostID, content)
	if err != nil {
		// En cas d'erreur, rediriger vers la page de connexion
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
