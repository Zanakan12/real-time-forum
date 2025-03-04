package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

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

func InitWebSocket() {
	go func() {
		for {
			msg := <-broadcast
			mutex.Lock()
			for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println("Erreur d'envoi WebSocket :", err)
					client.Close()
					delete(clients, client)
				}
			}
			mutex.Unlock()
		}
	}()

	// Vérification des inactifs
	go CheckInactiveUsers()
}

func CheckInactiveUsers() {
	for {
		mutex.Lock()
		for conn, username := range clients {
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				// L'utilisateur ne répond pas, on le déconnecte
				fmt.Println("Déconnexion pour inactivité :", username)
				conn.Close()
				delete(clients, conn)
				broadcast <- Message{Type: "user_list", Content: GetUserListJSON()}
			}
		}
		mutex.Unlock()
		// Vérifie toutes les 30 secondes
		time.Sleep(30 * time.Second)
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetCookie(w, r)
	fmt.Println("Connexion WebSocket de :", session.Username)

	userName, err := db.DecryptData(session.Username)
	if err != nil || userName == "" {
		fmt.Println("Refus de connexion WebSocket : Session invalide")
		http.Error(w, "Session invalide", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur WebSocket :", err)
		return
	}

	// Ajouter l'utilisateur à la liste des connectés
	mutex.Lock()
	clients[conn] = userName
	mutex.Unlock()

	// Notifier les autres utilisateurs
	broadcast <- Message{Type: "user_list", Content: GetUserListJSON()}

	// Nettoyage en cas de déconnexion
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		broadcast <- Message{Type: "user_list", Content: GetUserListJSON()}
		fmt.Println("Utilisateur déconnecté :", userName)
		conn.Close()
	}()

	// Lire les messages entrants
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Erreur WebSocket ou déconnexion détectée pour :", userName)
			break // Quitter la boucle et déclencher le `defer`
		}
		msg.Username = userName
		broadcast <- msg
	}
}

// GetUserListJSON retourne la liste des utilisateurs connectés en JSON
func GetUserListJSON() string {
	mutex.Lock()
	defer mutex.Unlock()
	usernames := []string{}
	for _, username := range clients {
		if !contains(usernames, username) {
			usernames = append(usernames, username)
		}
	}

	usersJSON, _ := json.Marshal(usernames)
	fmt.Println(string(usersJSON))
	return string(usersJSON)
}

// contains vérifie si un username est déjà dans la liste
func contains(slice []string, item string) bool {
	for _, val := range slice {
		if val == item {
			return true
		}
	}
	return false
}

// ChatPageHandler sert la page de discussion instantanée
func ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html") // Assurez-vous que chat.html est bien placé
}
