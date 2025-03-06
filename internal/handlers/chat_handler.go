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
	Type      string `json:"type"`     // "message" ou "user_list"
	Username  string `json:"username"` // Nom de l'utilisateur
	Content   string `json:"content"`  // Contenu du message
	Recipient string `json:"recipient"`
}

// Stockage des connexions WebSocket
var (
	clients   = make(map[*websocket.Conn]string) // Connexion → Username
	broadcast = make(chan WebSocketMessage, 10)  // Canal bufferisé pour éviter les blocages
	mutex     = sync.Mutex{}
)

// Initialisation du WebSocket (écoute des messages et gestion des inactifs)
func InitWebSocket() {
	go handleBroadcast()    // Goroutine pour gérer la diffusion des messages
	go checkInactiveUsers() // Goroutine pour détecter les utilisateurs inactifs
}

// Écoute des messages à diffuser à tous les clients
func handleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				fmt.Println("Erreur d'envoi WebSocket :", err)
				client.Close()
				delete(clients, client) // Suppression immédiate des connexions mortes
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
		inactiveClients := []*websocket.Conn{}

		for conn, username := range clients {
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fmt.Println("Déconnexion pour inactivité :", username)
				inactiveClients = append(inactiveClients, conn)
			}
		}

		// Supprime les connexions inactives
		for _, conn := range inactiveClients {
			delete(clients, conn)
			conn.Close()
		}

		// Mise à jour de la liste des utilisateurs actifs
		select {
		case broadcast <- WebSocketMessage{Type: "user_list", Content: GetUserListJSON()}:
		default:
			fmt.Println("Le canal de diffusion est plein, message ignoré.")
		}
		mutex.Unlock()
	}
}

// Gestion des connexions WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Vérification de la session utilisateur
	session := middlewares.GetCookie(w, r)
	userName, err := db.DecryptData(session.Username)
	if err != nil || userName == "" {
		fmt.Println("Refus de connexion WebSocket : Session invalide")
		http.Error(w, "Session invalide", http.StatusUnauthorized)
		return
	}

	fmt.Println("Connexion WebSocket de :", userName)

	// Mise à niveau de la connexion HTTP vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur WebSocket :", err)
		return
	}
	defer conn.Close() // Ferme la connexion à la fin de la fonction

	// Ajouter l'utilisateur à la liste des connectés
	mutex.Lock()
	clients[conn] = userName
	mutex.Unlock()

	// Mise à jour de la liste des utilisateurs connectés
	select {
	case broadcast <- WebSocketMessage{Type: "user_list", Content: GetUserListJSON()}:
	default:
		fmt.Println("Le canal de diffusion est plein, message ignoré.")
	}

	// Gestion des messages WebSocket
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Erreur lecture :", err)
			break
		}

		// Décodage du message JSON reçu
		var receivedMessage WebSocketMessage
		err = json.Unmarshal(msg, &receivedMessage)
		if err != nil {
			log.Println("Message invalide :", err)
			continue
		}

		log.Printf("Message reçu de %s: %s pour %s\n", receivedMessage.Username, receivedMessage.Content, receivedMessage.Recipient)
		db.SaveMessage(receivedMessage.Username, receivedMessage.Recipient, receivedMessage.Content)

		// Diffusion du message reçu à tous les clients
		select {
		case broadcast <- receivedMessage:
		default:
			fmt.Println("Le canal de diffusion est plein, message ignoré.")
		}
	}

	// Suppression de l'utilisateur après la déconnexion
	mutex.Lock()
	delete(clients, conn)
	mutex.Unlock()
	broadcast <- WebSocketMessage{Type: "user_list", Content: GetUserListJSON()}

	fmt.Println("Utilisateur déconnecté :", userName)
}

// Retourne la liste des utilisateurs connectés en JSON
func GetUserListJSON() string {
	mutex.Lock()
	defer mutex.Unlock()
	usernames := []string{}
	for _, username := range clients {
		if !contains(usernames, username) {
			usernames = append(usernames, username)
		}
	}

	usersJSON, err := json.Marshal(usernames)
	if err != nil {
		log.Println(err)
	}
	return string(usersJSON)
}

// Vérifie si un utilisateur est déjà dans la liste
func contains(slice []string, item string) bool {
	for _, val := range slice {
		if val == item {
			return true
		}
	}
	return false
}

// Servir la page de chat
func ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/pages/chat.html") // Assurez-vous que chat.html est bien placé
}
