package main

import (
	"adventmouton/bot"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		bot.Logformat(bot.ERROR, "No .env file found.\n")
		os.Exit(1)
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		bot.Logformat(bot.ERROR, "Missing DISCORD_TOKEN in .env or environment\n")
		os.Exit(2)
	}

	bot.Setup(token)
	bot.Loop()
}
