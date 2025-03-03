package config

const (
	WEBSITE_TITLE = "Golang Server Layout"
	DB_PATH       = "internal/database/forum.db"
	DB_USER       = "root"
	DB_PW         = "root"
)

var (
	IMG_EXT               = []string{".jpg", ".jpeg", ".gif", ".png"}
	ROLES                 = []string{"admin", "user", "moderator", "banned"}
	DISCORD_CLIENT_ID     = ""
	DISCORD_CLIENT_SECRET = ""
)
