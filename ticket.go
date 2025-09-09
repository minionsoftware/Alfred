package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func SendTicketEmbed(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "🎫 Create a Ticket",
		Description: "Click the button below to open a support ticket.",
		Color:       0x00b0f4,
	}

	button := discordgo.Button{
		Label:    "Open Ticket",
		Style:    discordgo.PrimaryButton,
		CustomID: "create_ticket",
	}

	msg := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{button},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: msg,
	})
}

func ShowTicketForm(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modal := &discordgo.InteractionResponseData{
		Title:    "Support Ticket",
		CustomID: "ticket_form",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "subject",
						Label:       "Subject",
						Style:       discordgo.TextInputShort,
						Placeholder: "Enter the issue subject...",
						Required:    true,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "description",
						Label:       "Description",
						Style:       discordgo.TextInputParagraph,
						Placeholder: "Describe your issue in detail...",
						Required:    true,
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: modal,
	})
}

func HandleFormSubmission(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *Config) {
	data := i.ModalSubmitData()
	subject := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	user := i.Member.User
	channelName := fmt.Sprintf("ticket-%s", user.Username)

	channel, err := s.GuildChannelCreateComplex(cfg.GuildId, discordgo.GuildChannelCreateData{
	Name:     channelName,
	Type:     discordgo.ChannelTypeGuildText,
	ParentID: cfg.TicketCategoryId,
	PermissionOverwrites: []*discordgo.PermissionOverwrite{
		{
			ID:   cfg.GuildId,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:    cfg.AdminRoleId,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages,
		},
		{
			ID:    user.ID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages,
		},
		{
			ID:    s.State.User.ID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages,
		},
	},
   })

	if err != nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	
	msg := &discordgo.MessageSend{
		Content: fmt.Sprintf(
			"📩 New ticket from <@%s>\n\n**Subject:** %s\n**Description:** %s\n",
			user.ID, subject, description,
		),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Close Ticket",
						Style:    discordgo.DangerButton,
						CustomID: "close_ticket",
					},
				},
			},
		},
	}
	
	_, err = s.ChannelMessageSendComplex(channel.ID, msg)
	if err != nil {
		fmt.Println("Error sending message in ticket channel:", err)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Your ticket has been created: <#%s>", channel.ID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func CloseTicket(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *Config) {
    isAdmin := false
    for _, roleID := range i.Member.Roles {
        if roleID == cfg.AdminRoleId {
            isAdmin = true
            break
        }
    }

    if !isAdmin {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "❌ You don't have permission to close tickets.",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }

    _, err := s.ChannelDelete(i.ChannelID)
    if err != nil {
        fmt.Println("Error deleting channel:", err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "⚠️ Failed to delete the ticket channel.",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "✅ Ticket channel deleted successfully.",
            Flags:   discordgo.MessageFlagsEphemeral,
        },
    })
}

