package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type Config struct {
	DBHost         string
	DBUserName     string
	DBUserPassword string
	DBName         string
	DBPort         int
	ServerPort     string
	DefaultOffset  string
	DefaultLimit   string
	API_SECRET     string
	BotToken       string
	Release_Mode   string
	WEBHOOK_URL    string
	ErrorBotToken  string // xatoliklar Telegram guruhga: ERROR_BOT_TOKEN, ERROR_CHAT_ID
	ErrorChatID    string
}

func LoadConfig() (config Config, err error) {
	err = godotenv.Load(".env")
	if err != nil {
		return
	}

	config = Config{
		DBHost:         cast.ToString(os.Getenv("DB_HOST")),
		DBUserName:     cast.ToString(os.Getenv("DB_USER")),
		DBUserPassword: cast.ToString(os.Getenv("DB_PASSWORD")),
		DBName:         cast.ToString(os.Getenv("DB_NAME")),
		DBPort:         cast.ToInt(os.Getenv("DB_PORT")),
		ServerPort:     cast.ToString(os.Getenv("PORT")),
		DefaultOffset:  cast.ToString(os.Getenv("DEFAULT_OFFSET")),
		DefaultLimit:   cast.ToString(os.Getenv("DEFAULT_LIMIT")),
		API_SECRET:     cast.ToString(os.Getenv("API_SECRET")),
		BotToken:       cast.ToString(os.Getenv("BOT_TOKEN")),
		Release_Mode:   cast.ToString(os.Getenv("RELEASE_MODE")),
		WEBHOOK_URL:    cast.ToString(os.Getenv("WEBHOOK_URL")),
		ErrorBotToken:  cast.ToString(os.Getenv("ERROR_BOT_TOKEN")),
		ErrorChatID:    cast.ToString(os.Getenv("ERROR_CHAT_ID")),
	}

	return
}
