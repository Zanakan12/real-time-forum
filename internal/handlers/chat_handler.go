package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"db"
	"middlewares"

	"github.com/gorilla/websocket"
)

// Upgrader WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Structure d'un message WebSocket
type Message struct {
	Type     string `json:"type"`     // "message" ou "user_list"
	Username string `json:"username"` // Nom de l'utilisateur
	Content  string `json:"content"`  // Contenu du message
}

// Stocker les connexions WebSocket
var clients = make(map[*websocket.Conn]string) // Connexion → Username
var broadcast = make(chan Message)
var mutex = sync.Mutex{}

// HandleWebSocket gère les connexions WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Récupérer la session utilisateur
	session := middlewares.GetCookie(w, r)
	if session.Username == "traveler" {
		fmt.Println("Refus de connexion WebSocket : Pas de session")
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}
	fmt.Println("Connexion WebSocket de :", session.Username)
	// Déchiffrer le nom d'utilisateur
	userName, err := db.DecryptData(session.Username)
	if err != nil || userName == "" {
		fmt.Println("Refus de connexion WebSocket : Session invalide")
		http.Error(w, "Session invalide", http.StatusUnauthorized)
		return
	}

	// Mise à niveau de la connexion HTTP en WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur WebSocket :", err)
		return
	}
	defer conn.Close()

	// Ajouter l'utilisateur à la liste des connectés
	mutex.Lock()
	clients[conn] = userName
	mutex.Unlock()

	// Notifier tous les utilisateurs
	broadcast <- Message{Type: "user_list", Content: getUserList()}

	// Lire les messages entrants
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Déconnexion de :", userName)

			// Supprimer l'utilisateur de la liste
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()

			// Notifier les autres utilisateurs
			broadcast <- Message{Type: "user_list", Content: getUserList()}
			break
		}
		msg.Username = userName // Associer l'utilisateur au message
		broadcast <- msg
	}
}

// getUserList génère une liste JSON des utilisateurs connectés
func getUserList() string {
	mutex.Lock()
	defer mutex.Unlock()
	usernames := []string{}
	for _, username := range clients {
		if !contains(usernames, username) {
			usernames = append(usernames, username)
		}
	}
	usersJSON, _ := json.Marshal(usernames)
	fmt.Println("Liste des utilisateurs connectés :", string(usersJSON))
	return string(usersJSON)
}

func contains(words []string, target string) bool {
	for _, word := range words {
		if word == target {
			return true
		}
	}
	return false
}

// ChatPageHandler sert la page de discussion instantanée
func ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html") // Assurez-vous que chat.html est dans un dossier "static"
}
