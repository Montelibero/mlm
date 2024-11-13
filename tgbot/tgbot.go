package tgbot

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/looplab/fsm"
)

type TGBot struct {
	l       *slog.Logger
	db      db.Querier
	bot     *tgbotapi.BotAPI
	stellar *stellar.Client
	distrib *distributor.Distributor
}

func (t *TGBot) Run(ctx context.Context) {
	updchan := t.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for {
		select {
		case upd := <-updchan:
			if upd.Message == nil {
				return
			}

			t.l.DebugContext(ctx, "[tg] new message",
				slog.Int64("from_id", upd.Message.From.ID),
				slog.Int64("chat_id", upd.Message.Chat.ID),
				slog.String("text", upd.Message.Text),
			)

			if err := t.handle(ctx, upd); err != nil {
				t.l.ErrorContext(ctx, "failed to handle update",
					slog.Int64("from_id", upd.Message.From.ID),
					slog.Int64("chat_id", upd.Message.Chat.ID),
					slog.String("text", upd.Message.Text),
					slog.String("error", err.Error()),
				)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (t *TGBot) handle(ctx context.Context, upd tgbotapi.Update) error {
	if !upd.Message.Chat.IsPrivate() || !upd.Message.IsCommand() {
		return nil
	}

	st, err := t.db.GetState(ctx, upd.Message.From.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		st = db.State{
			UserID: upd.Message.From.ID,
			State:  stateInit,
			Data:   make(map[string]interface{}),
			Meta:   make(map[string]interface{}),
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	sm := t.newSM()
	sm.SetState(st.State)

	st.Data["message"] = upd.Message.Text
	st.Data["message_id"] = upd.Message.MessageID
	st.Meta["username"] = upd.Message.From.UserName
	st.Meta["firstname"] = upd.Message.From.FirstName
	st.Meta["lastname"] = upd.Message.From.LastName
	st.Meta["chat_type"] = upd.Message.Chat.Type
	st.Meta["chat_title"] = upd.Message.Chat.Title
	st.Meta["chat_id"] = upd.Message.Chat.ID

	if err := sm.Event(ctx, upd.Message.Text, st); err != nil && !errors.Is(err, fsm.NoTransitionError{}) {
		return err
	}

	return nil
}

func New(
	l *slog.Logger,
	db db.Querier,
	bot *tgbotapi.BotAPI,
	stellar *stellar.Client,
	distrib *distributor.Distributor,
) *TGBot {
	return &TGBot{
		l:       l,
		db:      db,
		bot:     bot,
		stellar: stellar,
		distrib: distrib,
	}
}
