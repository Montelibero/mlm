package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	"github.com/Montelibero/mlm/tgbot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/stellar/go/clients/horizonclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := godotenv.Load(); err != nil {
		l.ErrorContext(ctx, err.Error())
	}

	cl := horizonclient.DefaultPublicNetClient

	pg, err := pgx.Connect(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	q := db.New(pg)

	stell := stellar.NewClient(cl)
	distrib := distributor.New(stell, q, pg)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	tgbot := tgbot.New(l, q, bot, distrib)

	tgbot.Run(ctx) // blocks
}
