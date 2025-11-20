package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (this is fine if running in production)")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Missing DISCORD_TOKEN in .env or environment")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(onInteraction)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Cannot open Discord session: %v", err)
	}

	// Register slash commands
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
		log.Fatalf("Cannot create slash command: %v", err)
	}

	log.Println("Bot is running. Press CTRL+C to exit.")

	// Wait for CTRL+C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	dg.Close()
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
