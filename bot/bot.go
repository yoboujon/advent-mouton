package bot

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var dg *discordgo.Session = nil
var filelist []string

func Setup(data EnvData) {
	var err error
	dg, err = discordgo.New("Bot " + data.DiscordToken)
	if err != nil {
		Logformat(ERROR, "Error creating Discord session: %v", err)
	}
	dg.AddHandler(onInteraction)
	err = dg.Open()
	if err != nil {
		Logformat(ERROR, "Cannot open Discord session: %v", err)
	}
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "advent",
		Description: "Advent related commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "help",
				Description: "Show help menu",
			},
		},
	})
	if err != nil {
		Logformat(ERROR, "Cannot create slash command: %v", err)
	}
	Logformat(INFO, "Downloading %s, this may take some time...\n", data.MediaName)
	filelist, err = downloadMedia(data.MediaURL, data.MediaName)
	if err != nil {
		Logformat(ERROR, "%s\n", err.Error())
	}
}

func Stop() {
	dg.Close()
}

func Loop() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	Logformat(INFO, "Bot is running. Press CTRL+C to exit.\n")
	for {
		select {
		case <-stop:
			print("\n")
			Logformat(INFO, "Stopping...\n")
			return
		default:
			time.Sleep(time.Second)
			botLogic()
		}
	}
}

func onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "advent" {
		// Check which subcommand
		options := i.ApplicationCommandData().Options
		if len(options) > 0 && options[0].Name == "help" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Advent Help Menu**\n\n• `/advent help` — Show this help menu",
				},
			})
		}
	}
}

func downloadMedia(url string, zipName string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out, err := os.Create(zipName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return nil, err
	}

	// unzipping files
	result, err := unzip(zipName)
	if err != nil {
		return nil, err
	}
	os.Remove(zipName)

	return result, nil
}

func unzip(zipPath string) ([]string, error) {
	var fileMap []string
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		// Creating "assets" folder
		err = os.MkdirAll("assets", 0755)
		path := filepath.Join("assets", filepath.Base(f.Name))
		if err != nil {
			rc.Close()
			return nil, err
		}

		// Checking if in zip file is a gif/mp4/webp
		isMedia := filepath.Ext(f.Name) == ".gif" || filepath.Ext(f.Name) == ".mp4" || filepath.Ext(f.Name) == ".webm"
		if !isMedia {
			rc.Close()
			continue
		}

		// Creating file and extracting it.
		print("\r")
		Logformat(INFO, "Extracting '%s'...", f.Name)
		outFile, err := os.Create(path)
		if err != nil {
			rc.Close()
			return nil, err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return nil, err
		}

		fileMap = append(fileMap, path)
	}
	print("\n")

	return fileMap, nil
}

func botLogic() {
	now := time.Now()
	if now.Month() != time.December {
		return
	}

}
