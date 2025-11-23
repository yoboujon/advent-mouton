package bot

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type EnvData struct {
	DiscordToken string
	Msg          string
	MediaURL     string
	MediaName    string
}

var env_data EnvData

func GetData() (EnvData, error) {
	required := []string{"DISCORD_TOKEN", "MSG", "MEDIA_URL", "MEDIA_NAME"}
	values := make(map[string]string)
	err := godotenv.Load()
	if err != nil {
		return EnvData{}, err
	}

	for _, key := range required {
		v := os.Getenv(key)
		if v == "" {
			return EnvData{}, fmt.Errorf("missing %s in .env or environment", key)
		}
		values[key] = v
	}

	env_data = EnvData{
		DiscordToken: values["DISCORD_TOKEN"],
		Msg:          values["MSG"],
		MediaURL:     values["MEDIA_URL"],
		MediaName:    values["MEDIA_NAME"],
	}
	return env_data, nil
}

func GetRawData() EnvData {
	return env_data
}
