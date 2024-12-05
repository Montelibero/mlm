package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AllowedUserIDs map[int64]struct{}
	PostgresDSN    string
	TelegramToken  string
	Address        string
	Seed           string
}

func Get() *Config {
	_ = godotenv.Load()

	allowedUserIDs := make(map[int64]struct{})

	for _, idStr := range strings.Split(os.Getenv("ALLOWED_USER_IDS"), ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}

		allowedUserIDs[id] = struct{}{}
	}

	return &Config{
		PostgresDSN:    os.Getenv("POSTGRES_DSN"),
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		Address:        os.Getenv("STELLAR_ADDRESS"),
		Seed:           os.Getenv("STELLAR_SEED"),
		AllowedUserIDs: allowedUserIDs,
	}
}
