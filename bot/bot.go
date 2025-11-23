package bot

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var dg *discordgo.Session = nil
var filelist []string
var bot_data EnvData

type MediaMessage struct {
	Timestamp time.Time
	FileName  string
}

func Setup(data EnvData) {
	var err error
	bot_data = data
	dg, err = discordgo.New("Bot " + bot_data.DiscordToken)
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
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Update the media folder",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setup",
				Description: "Setup the channel to send media",
			},
		},
	})
	if err != nil {
		Logformat(ERROR, "Cannot create slash command: %v", err)
	}
	filelist, err = downloadMedia(bot_data.MediaURL, bot_data.MediaName)
	if err != nil {
		Logformat(ERROR, "%s\n", err.Error())
	}
	err = loadGuildChannels()
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
			time.Sleep(time.Duration(time.Second * 1))
			botLogic()
		}
	}
}

func onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	if i.ApplicationCommandData().Name == "advent" {
		// Check which subcommand
		options := i.ApplicationCommandData().Options
		if len(options) > 0 && options[0].Name == "help" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "**Advent Help Menu**\n\n• `/advent help` — Show this help menu\n• `/advent update` — Update the media",
				},
			})
		}
		if len(options) > 0 && options[0].Name == "update" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Updating media, please wait...",
				},
			})

			filelist, err = downloadMedia(bot_data.MediaURL, bot_data.MediaName)
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Failed to update media ! :c",
				})
			} else {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Updated media with **%d files** ! c:", len(filelist)),
				})
			}
		}
		if len(options) > 0 && options[0].Name == "setup" {
			addGuildChannels(i.GuildID, i.ChannelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Selected channel to send media!",
				},
			})
		}
	}
}

func downloadMedia(url string, zipName string) ([]string, error) {
	Logformat(INFO, "Downloading %s, this may take some time...\n", zipName)
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

		fileMap = append(fileMap, strings.TrimPrefix(path, "assets/"))
	}
	print("\n")

	return fileMap, nil
}

func botLogic() {
	now := time.Now()
	for guildID, channelID := range guildChannels {
		msg, err := getBotMediaMessages(dg, channelID, dg.State.User.ID)
		if err != nil {
			Logformat(WARNING, "Failed to get messages from server %s: %s", guildID, err.Error())
			return
		}

		send := true
		for _, element := range msg {
			timestamp := element.Timestamp
			if (timestamp.Day() == now.Day()) && (timestamp.Month() == now.Month()) && (timestamp.Year() == now.Year()) {
				send = false
			}
			if now.Hour() < 8 || (now.Month() != time.December) {
				send = false
			}
		}
		if send {
			fileName, err := getRandomFile(msg)
			if err != nil {
				Logformat(WARNING, "Failed to send to server %s: %s", guildID, err.Error())
			}
			_, err = dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Content: env_data.Msg,
				Files: []*discordgo.File{
					{
						Name:   fileName,
						Reader: mustOpen("assets/" + fileName),
					},
				},
			})
			if err != nil {
				Logformat(WARNING, "Failed to send to server %s: %s", guildID, err.Error())
			}
		}
	}
}

func getBotMediaMessages(s *discordgo.Session, channelID, botID string) ([]MediaMessage, error) {
	var mediaMessages []MediaMessage
	beforeID := ""

	for {
		msgs, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}

		for _, m := range msgs {
			// If its our bot sending the message
			if m.Author.ID != botID {
				continue
			}
			// Checking there is a media
			if len(m.Attachments) == 0 {
				continue
			}

			for _, a := range m.Attachments {
				mediaMessages = append(mediaMessages, MediaMessage{
					Timestamp: m.Timestamp,
					FileName:  a.Filename,
				})
			}
		}

		beforeID = msgs[len(msgs)-1].ID
	}

	return mediaMessages, nil
}

func getRandomFile(msg []MediaMessage) (string, error) {
	sent := make(map[string]bool)
	for _, f := range filelist {
		f = filepath.Base(f)
		sent[f] = false
		for _, m := range msg {
			if m.FileName == f {
				sent[f] = true
			}
		}
	}

	// Checking if all media are already sent
	allSent := true
	for _, wasSent := range sent {
		if !wasSent {
			allSent = false
			break
		}
	}
	if allSent {
		return "", errors.New("all media already sent")
	}

	// For loop to find a message that was not sent
	for {
		randomFile := filelist[rand.Intn(len(filelist))]
		if !sent[filepath.Base(randomFile)] {
			return randomFile, nil
		}
	}
}

func mustOpen(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}
