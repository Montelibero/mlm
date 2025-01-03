package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AllowedUserIDs          map[int64]struct{}
	PostgresDSN             string
	TelegramToken           string
	Address                 string
	Seed                    string
	Submit                  bool
	WithoutReport           bool
	ReportToChatID          int64
	ReportToMessageThreadID int64
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

	reportToChatID, _ := strconv.ParseInt(os.Getenv("REPORT_TO_CHAT_ID"), 10, 64)
	reportToMessageThreadID, _ := strconv.ParseInt(os.Getenv("REPORT_TO_MESSAGE_THREAD_ID"), 10, 64)

	return &Config{
		PostgresDSN:             os.Getenv("POSTGRES_DSN"),
		TelegramToken:           os.Getenv("TELEGRAM_TOKEN"),
		Address:                 os.Getenv("STELLAR_ADDRESS"),
		Seed:                    os.Getenv("STELLAR_SEED"),
		Submit:                  os.Getenv("SUBMIT") == "true",
		WithoutReport:           os.Getenv("WITHOUT_REPORT") == "true",
		ReportToChatID:          reportToChatID,
		ReportToMessageThreadID: reportToMessageThreadID,
		AllowedUserIDs:          allowedUserIDs,
	}
}
