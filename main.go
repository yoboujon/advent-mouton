package main

import (
	"adventmouton/bot"
	"os"
)

func main() {
	env_data, err := bot.GetData()
	if err != nil {
		bot.Logformat(bot.ERROR, "%s\n", err.Error())
		os.Exit(1)
	}
	bot.Setup(env_data)
	bot.Loop()
}
