package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var dg *discordgo.Session = nil

func Setup(token string) {
	var err error
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}
	dg.AddHandler(onInteraction)
	err = dg.Open()
	if err != nil {
		log.Fatalf("Cannot open Discord session: %v", err)
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
		log.Fatalf("Cannot create slash command: %v", err)
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

func botLogic() {

}
