package tgbot

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Montelibero/mlm/config"
	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/stellar"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgtype"
)

const QuerySubmitXDR = "submit_xdr_"

type TGBot struct {
	cfg     *config.Config
	l       *slog.Logger
	bot     *bot.Bot
	q       *db.Queries
	stellar *stellar.Client
}

func (t *TGBot) Run(ctx context.Context) {
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, QuerySubmitXDR, bot.MatchTypePrefix, t.handleSubmitXDR)

	t.bot.Start(ctx)
}

func (t *TGBot) handleSubmitXDR(ctx context.Context, _ *bot.Bot, upd *models.Update) {
	if upd.CallbackQuery == nil {
		return
	}

	if _, ok := t.cfg.AllowedUserIDs[upd.CallbackQuery.From.ID]; !ok {
		_, _ = t.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: upd.CallbackQuery.ID,
			Text:            "Только админ может жать кнопки",
			ShowAlert:       true,
		})

		return
	}

	reportIDStr := strings.ReplaceAll(upd.CallbackQuery.Data, QuerySubmitXDR, "")
	reportID, _ := strconv.ParseInt(reportIDStr, 10, 64)
	if reportID == 0 {
		return
	}

	l := t.l.With(slog.Int64("report_id", reportID))

	rep, err := t.q.GetReport(ctx, reportID)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		return
	}

	hash, err := t.stellar.SubmitXDR(ctx, t.cfg.Seed, rep.Xdr)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		return
	}

	if err := t.q.SetReportHash(ctx, db.SetReportHashParams{
		Hash:     pgtype.Text{String: hash, Valid: true},
		ReportID: reportID,
	}); err != nil {
		l.ErrorContext(ctx, err.Error())
		return
	}
}

func New(cfg *config.Config, l *slog.Logger, b *bot.Bot, q *db.Queries, s *stellar.Client) *TGBot {
	return &TGBot{
		cfg:     cfg,
		l:       l,
		bot:     b,
		q:       q,
		stellar: s,
	}
}
