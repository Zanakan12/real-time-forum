package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Configuration de l'upgrader WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepter toutes les connexions (à sécuriser en prod)
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var mutex = sync.Mutex{}

// Structure d'un message
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// HandleWebSocket gère la connexion WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Mise à niveau de la connexion HTTP en WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur WebSocket :", err)
		return
	}
	defer conn.Close()

	// Ajouter le client
	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	// Lire les messages en boucle
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Déconnexion WebSocket :", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
		broadcast <- msg
	}
}

// BroadcastMessages envoie les messages à tous les clients connectés
func BroadcastMessages() {
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
}

// ChatPageHandler sert la page de discussion instantanée
func ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html") // Assurez-vous que chat.html est dans un dossier "static"
}
