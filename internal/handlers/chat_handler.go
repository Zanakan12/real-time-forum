package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
type WebSocketMessage struct {
	Type      string `json:"type"`      // "message" ou "user_list"
	Username  string `json:"username"`  // Expéditeur
	Recipient string `json:"recipient"` // Destinataire
	Content   string `json:"content"`   // Contenu du message
	Read      bool   `json:"read"`      // Indique si le message a été lu
	CreatedAt string `json:"created_at"`
}

// Stockage des connexions WebSocket (map username -> websocket.Conn)
var (
	clients   = make(map[string]*websocket.Conn) // Connexions des utilisateurs
	broadcast = make(chan WebSocketMessage, 10)  // Canal bufferisé pour diffusion des messages
	mutex     = sync.Mutex{}
)

// Initialisation du WebSocket (écoute des messages et gestion des inactifs)
func InitWebSocket() {
	go handleBroadcast()
	go checkInactiveUsers()
}

// Diffusion des messages à tous les utilisateurs (utile pour mise à jour de la liste des connectés)
func handleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for _, conn := range clients {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("Erreur d'envoi WebSocket :", err)
				conn.Close()
				delete(clients, msg.Username)
			}
		}
		mutex.Unlock()
	}
}

// Vérifie les utilisateurs inactifs et les déconnecte
func checkInactiveUsers() {
	for {
		time.Sleep(30 * time.Second)

		mutex.Lock()
		for username, conn := range clients {
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fmt.Println("Déconnexion pour inactivité :", username)
				delete(clients, username)
				conn.Close()
			}
		}
		broadcast <- WebSocketMessage{Type: "user_list", Content: GetUserListJSON()}
		mutex.Unlock()
	}
}

// Gestion des connexions WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Vérification de la session utilisateur
	session := middlewares.GetCookie(w, r)
	userName, err := db.DecryptData(session.Username)
	if err != nil || userName == "" {
		http.Error(w, "Session invalide", http.StatusUnauthorized)
		return
	}

	// Upgrade vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erreur WebSocket :", err)
		return
	}

	// Ajouter l'utilisateur à la liste des connectés
	mutex.Lock()
	clients[userName] = conn
	mutex.Unlock()

	log.Println("Utilisateur connecté :", userName)

	// Envoyer les messages non lus à la reconnexion
	fetchUnreadMessages(userName)

	// Mettre à jour la liste des utilisateurs connectés
	updateUserList()

	// Gestion des messages WebSocket
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Déconnexion :", userName)
			break
		}

		// Décoder le message reçu
		var receivedMessage WebSocketMessage
		if err := json.Unmarshal(msg, &receivedMessage); err != nil {
			log.Println("Message invalide :", err)
			continue
		}

		log.Printf("Message de %s → %s : %s\n", receivedMessage.Username, receivedMessage.Recipient, receivedMessage.Content)

		// Sauvegarde et envoi du message
		sendMessageToUser(receivedMessage.Recipient, receivedMessage)
	}

	// Suppression de l'utilisateur après la déconnexion
	mutex.Lock()
	delete(clients, userName)
	mutex.Unlock()
	updateUserList()

	conn.Close()
}

// Envoi d'un message à un utilisateur spécifique
func sendMessageToUser(toUsername string, message WebSocketMessage) {
	mutex.Lock()
	conn, online := clients[toUsername]
	mutex.Unlock()

	if online {
		// Utilisateur en ligne → Envoi direct
		err := conn.WriteJSON(message)
		if err != nil {
			log.Printf("Erreur d'envoi à %s: %v", toUsername, err)
			mutex.Lock()
			delete(clients, toUsername) // Supprimer la connexion si elle est cassée
			mutex.Unlock()
		}
		message.Read = true
		db.SaveMessage(message.Username, message.Recipient, message.Content, message.Read)
	} else {
		// Utilisateur hors ligne → Stockage en base
		message.Read = false
		db.SaveMessage(message.Username, message.Recipient, message.Content, message.Read)
		log.Printf("Utilisateur %s hors ligne, message stocké.", toUsername)
	}
}

// Récupère les messages non lus pour un utilisateur et les lui envoie
func fetchUnreadMessages(username string) {
	messages := db.GetUnreadMessages(username)

	mutex.Lock()
	conn, online := clients[username]
	mutex.Unlock()

	if online {
		for _, msg := range messages {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Erreur d'envoi du message stocké à %s: %v", username, err)
			} else {
				db.MarkMessageAsRead(msg)
			}
		}
	}
}

// Mise à jour de la liste des utilisateurs connectés
func updateUserList() {
	userList := GetUserListJSON()
	broadcast <- WebSocketMessage{Type: "user_list", Content: userList}
}

// Retourne la liste des utilisateurs connectés en JSON
func GetUserListJSON() string {
	mutex.Lock()
	defer mutex.Unlock()
	usernames := []string{}
	for username := range clients {
		usernames = append(usernames, username)
	}
	usersJSON, _ := json.Marshal(usernames)
	return string(usersJSON)
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html")
}
