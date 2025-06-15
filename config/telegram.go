package config

type TeleConfig struct {
	BotToken string `envconfig:"TELEGRAM_BOT_TOKEN"`
	ChatID   string `envconfig:"TELEGRAM_CHAT_ID"`
}
