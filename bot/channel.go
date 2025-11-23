package bot

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	guildChannels = make(map[string]string)
	mu            sync.RWMutex
)

func loadGuildChannels() error {
	mu.Lock()
	defer mu.Unlock()

	Logformat(INFO, "Loading channels...\n")
	file, err := os.OpenFile("channel.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&guildChannels)
	if err != nil {
		Logformat(WARNING, "No channel found in any server.\n")
	}
	return nil
}

func saveGuildChannels() error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create("channel.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(&guildChannels)
}

func addGuildChannels(GuildID string, ChannelID string) {
	guildChannels[GuildID] = ChannelID
	saveGuildChannels()
}
