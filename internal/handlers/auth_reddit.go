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
	RedditOauthConfig = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "https://localhost:8080/reddit-callback",
		Scopes:       []string{"identity"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.reddit.com/api/v1/authorize",
			TokenURL: "https://www.reddit.com/api/v1/access_token",
		},
	}
	oauthStateStringReddit = "random"
)

func HandleRedditLogin(w http.ResponseWriter, r *http.Request) {
	RedditOauthConfig.ClientID = os.Getenv("REDDIT_CLIENT_ID")
	RedditOauthConfig.ClientSecret = os.Getenv("REDDIT_CLIENT_SECRET")
	url := RedditOauthConfig.AuthCodeURL(oauthStateStringReddit)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleRedditCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateStringReddit {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := RedditOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Échec de l'échange du code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := RedditOauthConfig.Client(r.Context(), token)
	req, err := http.NewRequest("GET", "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		http.Error(w, "Erreur lors de la création de la requête: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "MonApp/1.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erreur lors de la requête à Reddit: %v", err)
		http.Error(w, "Échec de la récupération des informations utilisateur: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Réponse de Reddit: %s", string(body))

	var user map[string]interface{}
	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Échec du décodage des informations utilisateur: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Utiliser le nom d'utilisateur Reddit comme email (Reddit ne fournit pas l'email)
	email := user["name"].(string) + "@reddit.com"

	// Check if the user exists
	authUser, logOAuthErr := db.UserSelectLoginOAuth(email)
	if logOAuthErr != nil {
		log.Printf("Utilisateur non trouvé, création d'un nouveau compte pour: %s", email)

		// Reddit fournit directement le nom d'utilisateur, pas besoin de vérifier
		username := user["name"].(string)

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

	middlewares.CreateSession(w, authUser.ID, authUser.Username, authUser.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
