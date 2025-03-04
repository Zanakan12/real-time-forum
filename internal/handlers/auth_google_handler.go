package handlers

import (
	"crypto/rand"
	"db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"middlewares"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars   = "0123456789"
	specialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?/`~"
	allChars     = lowerChars + upperChars + digitChars + specialChars
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://localhost:8080/google-callback",
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
	oauthStateStringGoogle = "random"
)

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	url := googleOauthConfig.AuthCodeURL(oauthStateStringGoogle)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateStringGoogle {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Échec de l'échange du code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(r.Context(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "Échec de la récupération des informations utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Échec de la lecture des informations utilisateur: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var user map[string]interface{}
	if err := json.Unmarshal(contents, &user); err != nil {
		http.Error(w, "Échec du décodage des informations utilisateur: "+err.Error(), http.StatusInternalServerError)
		return
	}

	email, ok := user["email"].(string)
	if !ok {
		http.Error(w, "Email de l'utilisateur non trouvé", http.StatusInternalServerError)
		return
	}

	// Check if the user exists
	authUser, logOAuthErr := db.UserSelectLoginOAuth(email)
	log.Println("authUser", authUser)
	if logOAuthErr != nil {
		log.Printf("Utilisateur non trouvé, création d'un nouveau compte pour: %s", email)
		username, ok := user["name"].(string)
		if !ok { // Use email as username if name is not available
			username = email[:strings.Index(email, "@")] // Remove domain from email
		}
		password, err := GeneratePassword(12)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		err = db.UserInsertRegister(email, username, password, "user")
		if err != nil {
			log.Printf("Erreur lors de la création de l'utilisateur: %v", err)
			http.Error(w, "Échec de la création de l'utilisateur: "+err.Error(), http.StatusInternalServerError)
			return
		}

		authUser, err = db.UserSelectLoginOAuth(email)
		if err != nil {
			log.Printf("Erreur lors de la récupération de l'utilisateur après création: %v", err)
			http.Error(w, "Échec de la récupération de l'utilisateur après création: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Utilisateur authentifié: ID=%d, Username=%s, Role=%s", authUser.ID, authUser.Username, authUser.Role)
	authUser.Username, _ = db.DecryptData(authUser.Username)
	// Create session using your custom CreateSession function
	middlewares.CreateSession(w, authUser.ID, authUser.Username, authUser.Role)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GeneratePassword génère un mot de passe robuste
func GeneratePassword(length int) (string, error) {
	if length < 8 { // Minimum recommandé : 8 caractères
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	password := make([]byte, length)

	// S'assurer que le mot de passe contient au moins un caractère de chaque type
	charSets := []string{lowerChars, upperChars, digitChars, specialChars}
	for i, set := range charSets {
		char, err := randomChar(set)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Remplir le reste du mot de passe avec des caractères aléatoires
	for i := len(charSets); i < length; i++ {
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Mélanger le mot de passe pour éviter une prévisibilité
	shuffle(password)

	return string(password), nil
}

// randomChar sélectionne un caractère aléatoire d'une chaîne
func randomChar(charSet string) (byte, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
	if err != nil {
		return 0, err
	}
	return charSet[index.Int64()], nil
}

// shuffle mélange un tableau d'octets de manière aléatoire
func shuffle(password []byte) {
	for i := range password {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(len(password))))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
}
