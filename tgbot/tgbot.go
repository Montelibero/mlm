package tgbot

import (
	"context"
	"log/slog"

	"github.com/Montelibero/mlm/config"
)

type TGBot struct {
	cfg *config.Config
	l   *slog.Logger
}

func (t *TGBot) Run(ctx context.Context) {
}
