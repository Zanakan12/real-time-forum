package main

import (
	"io/ioutil"
	"os"
	"server"
	"strings"
)

func init() {
	content, err := ioutil.ReadFile(".env")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "AES_KEY":
			os.Setenv("AES_KEY", value)
		case "GOOGLE_CLIENT_ID":
			os.Setenv("GOOGLE_CLIENT_ID", value)
		case "GOOGLE_CLIENT_SECRET":
			os.Setenv("GOOGLE_CLIENT_SECRET", value)
		case "REDDIT_CLIENT_ID":
			os.Setenv("REDDIT_CLIENT_ID", value)
		case "REDDIT_CLIENT_SECRET":
			os.Setenv("REDDIT_CLIENT_SECRET", value)
		case "DISCORD_CLIENT_ID":
			os.Setenv("DISCORD_CLIENT_ID", value)
		case "DISCORD_CLIENT_SECRET":
			os.Setenv("DISCORD_CLIENT_SECRET", value)
		}
	}
}

func main() {
	server.InitServer()
}
