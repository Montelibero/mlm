package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Montelibero/mlm"
	"github.com/Montelibero/mlm/config"
	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stellar/go/clients/horizonclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := config.Get()

	goose.SetDialect("pgx")
	goose.SetBaseFS(mlm.EmbedMigrations)

	conn, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	if err := goose.UpContext(ctx, conn, "migrations"); err != nil { //
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	if err := goose.VersionContext(ctx, conn, "migrations"); err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	cl := horizonclient.DefaultPublicNetClient

	pg, err := pgx.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	defer pg.Close(ctx)

	q := db.New(pg)

	stell := stellar.NewClient(cl)
	distrib := distributor.New(cfg, stell, q, pg)

	_ = distrib
}
