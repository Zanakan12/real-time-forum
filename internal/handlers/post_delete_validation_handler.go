package handlers

import (
	"db"
	"net/http"
	"strconv"
)

func PostDeleteValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode est POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Récupérer les données du formulaire
	postID := r.FormValue("post_id")
	convPostID, _ := strconv.Atoi(postID)

	// Appeler la fonction de validation dans la base de données
	err := db.PostDelete(convPostID)
	if err != nil {
		// En cas d'erreur, rediriger vers la page de connexion
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
