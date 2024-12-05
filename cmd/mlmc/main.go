package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Montelibero/mlm/config"
	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	"github.com/jackc/pgx/v5"
	"github.com/stellar/go/clients/horizonclient"
)

func main() {
	ctx := context.Background()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).WithGroup("mlmc")

	cfg := config.Get()

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

	res, err := distrib.Distribute(ctx)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	l = l.With(slog.Int64("report_id", res.ReportID))

	l.InfoContext(ctx, "report done",
		slog.Int("conflicts", len(res.Conflicts)),
		slog.Int("distributes", len(res.Distributes)),
		slog.Int("recommends", len(res.Recommends)),
	)

	if !cfg.Submit {
		return
	}

	hash, err := stell.SubmitXDR(ctx, cfg.Seed, res.XDR)
	if err != nil {
		l.ErrorContext(ctx, "failed to submit xdr")
		os.Exit(1)
	}

	l.InfoContext(ctx, "report xdr submitted",
		slog.String("hash", hash))
}
