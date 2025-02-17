package config

import (
	"os"
)

type AppConfig struct {
	TelegramBotToken string
}

func NewConfig() *AppConfig {
	return &AppConfig{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}
}
