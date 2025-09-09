package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
			os.Exit(1)
		}
		configPath = filepath.Join(cwd, "config.json")
	}

	cfg, err := ReadConfig(configPath)
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandlerOnce(ready)
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	    interactionCreate(s, i, cfg)
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	registerCommands(dg, cfg)

	fmt.Println("Bot is now running.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	fmt.Println("Shutting down.")
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Bot is ready.")
}

func registerCommands(s *discordgo.Session, cfg *Config) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "ticket-setup",
		Description: "Send ticket creation embed with button",
	}

	s.ApplicationCommandCreate(s.State.User.ID, cfg.GuildId, cmd)
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *Config) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name == "ticket-setup" {
			SendTicketEmbed(s, i)
		}
	case discordgo.InteractionMessageComponent:
		    switch i.MessageComponentData().CustomID {
    case "create_ticket":
        ShowTicketForm(s, i)
    case "close_ticket":
        CloseTicket(s, i, cfg)
    }
	case discordgo.InteractionModalSubmit:
		if i.ModalSubmitData().CustomID == "ticket_form" {
			HandleFormSubmission(s, i, cfg)
		}
	}
}

