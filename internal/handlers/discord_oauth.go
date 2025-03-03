package handlers

import (
	"db"
	"encoding/json"
	"io/ioutil"
	"middlewares"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	discordOauthConfig = &oauth2.Config{
		RedirectURL:  "https://localhost:8080/dis-callback",
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	oauthStateString = "random"
)

func HandleDiscordLogin(w http.ResponseWriter, r *http.Request) {
	discordOauthConfig.ClientID = os.Getenv("DISCORD_CLIENT_ID")
	discordOauthConfig.ClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	url := discordOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := discordOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	client := discordOauthConfig.Client(r.Context(), token)
	response, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	var user map[string]interface{}
	if err := json.Unmarshal(contents, &user); err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	//Chek if the user exists
	authUser, logOAuthErr := db.UserSelectLoginOAuth(user["email"].(string))
	if logOAuthErr != nil {
		err := db.UserInsertRegisterOAuth(user["email"].(string), user["username"].(string), "user")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		authUser, err = db.UserSelectLoginOAuth(user["email"].(string))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
	}
	middlewares.CreateSession(w, authUser.ID, authUser.Username, authUser.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
