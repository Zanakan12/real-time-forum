package handlers

import (
	"db"
	"encoding/json"
	"io/ioutil"
	"log"
	"middlewares"
	"net/http"
	"os"

	"golang.org/x/oauth2"
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
		http.Error(w, "Échec de la récupération des informations utilisateur: "+err.Error(), http.StatusInternalServerError)
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
	if logOAuthErr != nil {
		log.Printf("Utilisateur non trouvé, création d'un nouveau compte pour: %s", email)
		username, ok := user["name"].(string)
		if !ok {
			username = email // Use email as username if name is not available
		}
		err := db.UserInsertRegisterOAuth(email, username, "user")
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

	// Create session using your custom CreateSession function
	middlewares.CreateSession(w, authUser.ID, authUser.Username, authUser.Role)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
