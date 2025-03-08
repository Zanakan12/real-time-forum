package handlers

import (
	"db"
	"encoding/json"
	"fmt"
	"log"
	"middlewares"
	"net/http"
)

// GetUserHandler retourne l'utilisateur connecté
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	userName, err := db.DecryptData(session.Username)
	if err != nil {
		fmt.Println("Erreur lors du décryptage de l'user")
	}
	user := struct {
		Username string `json:"username"`
	}{
		Username: userName,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Println("Erreur lors de l'encodage JSON:", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
	}
}

// GetUserListHandler retourne la liste des utilisateurs connectés en JSON
func GetUserListHandler(w http.ResponseWriter, r *http.Request) {
	usernames := GetUserListJSON()

	// Configuration de la réponse HTTP
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encodage de la liste des utilisateurs en JSON et envoi de la réponse
	if err := json.NewEncoder(w).Encode(usernames); err != nil {
		http.Error(w, "Erreur lors de la génération du JSON", http.StatusInternalServerError)
		fmt.Println("Erreur JSON:", err)
	}
}

func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)

	recipient := r.URL.Query().Get("recipient")
	userName, err := db.DecryptData(session.Username)

	if err != nil {
		http.Error(w, "Erreur lors de la récupération du décrytpage de l'username", http.StatusInternalServerError)
		return
	}
	fmt.Println("message debug")
	messages, err := db.GetMessages(userName, recipient)

	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages", http.StatusInternalServerError)
		return
	}
	usersJSON, _ := json.Marshal(messages)

	fmt.Println("message", messages)
	json.NewEncoder(w).Encode(string(usersJSON))
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := db.GetAllUser()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
		return
	}

	for i := range users {
		var decryptErr error
		users[i].Username, decryptErr = db.DecryptData(users[i].Username)
		if decryptErr != nil {
			log.Println("Erreur lors du décryptage des usernames :", decryptErr)
		}
	}

	// Convertir la liste des utilisateurs en JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
