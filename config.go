package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	AdminRoleId string `json:"admin_role_id"`
	Token string `json:"token"`
	GuildId string `json:"guild_id"`
	TicketCategoryId string `json:"ticket_category_id"`
}

func ReadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payload Config
	err = json.Unmarshal(file, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}
