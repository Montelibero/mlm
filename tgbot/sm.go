package tgbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/Montelibero/mlm"
	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
	"github.com/samber/lo"
	"github.com/stellar/go/txnbuild"
)

const (
	// states
	stateInit   = "state_init"
	stateMain   = "state_main"
	stateResult = "state_result"

	// events
	eventStart     = "/start"
	eventReports   = "/reports"
	eventMLMDryRun = "/mlm_dry_run"
)

var mainButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(eventReports),
		tgbotapi.NewKeyboardButton(eventMLMDryRun),
	),
)

var resultButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(eventStart),
	),
)

func (t *TGBot) newSM() *fsm.FSM {
	return fsm.NewFSM(
		stateInit,
		fsm.Events{
			{Name: eventStart, Src: []string{stateInit}, Dst: stateMain},
			{Name: eventStart, Src: []string{stateResult}, Dst: stateMain},
			{Name: eventMLMDryRun, Src: []string{stateMain}, Dst: stateResult},
			{Name: eventMLMDryRun, Src: []string{stateResult}, Dst: stateResult},
		},
		fsm.Callbacks{
			eventStart: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				if err := t.db.CreateState(ctx, db.CreateStateParams{
					UserID: st.UserID,
					State:  stateMain,
					Data:   st.Data,
					Meta:   st.Meta,
				}); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, "Welcome to MLM bot")
				msg.ReplyMarkup = mainButtons
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},

			eventMLMDryRun: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				if err := t.db.CreateState(ctx, db.CreateStateParams{
					UserID: st.UserID,
					State:  stateResult,
					Data:   st.Data,
					Meta:   st.Meta,
				}); err != nil {
					e.Cancel(err)
					return
				}

				if _, err := t.bot.Send(tgbotapi.NewMessage(st.UserID, "Processing...")); err != nil {
					e.Cancel(err)
					return
				}

				res, err := t.distrib.Distribute(ctx)
				if err != nil {
					e.Cancel(err)
					return
				}

				accountDetail, err := t.stellar.AccountDetail(distributor.Account)
				if err != nil {
					e.Cancel(err)
					return
				}

				ops := lo.Map(res.Distributes, func(d mlm.Distribute, _ int) txnbuild.Operation {
					return &txnbuild.Payment{
						Destination: d.AccountID,
						Amount:      fmt.Sprintf("%.7f", d.Amount),
						Asset:       txnbuild.CreditAsset{Code: stellar.EURMTLAsset, Issuer: stellar.EURMTLIssuer},
					}
				})

				tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
					SourceAccount:        &accountDetail,
					IncrementSequenceNum: true,
					Operations:           ops,
					BaseFee:              10000,
					Memo:                 txnbuild.MemoText("mtla mlm distribution"),
					Preconditions: txnbuild.Preconditions{
						TimeBounds: txnbuild.NewInfiniteTimeout(),
					},
				})
				if err != nil {
					e.Cancel(err)
					return
				}

				xdr, err := tx.Base64()
				if err != nil {
					e.Cancel(err)
					return
				}

				// TODO(xdefrag): save report

				reportStr := &strings.Builder{}
				fmt.Fprintf(reportStr, "<b>Report results</b>\n")

				for _, distrib := range res.Distributes {
					fmt.Fprintf(reportStr, "  <b>AccountID</b>: <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n <b>Amount</b>: %f EURMTL\n\n", distrib.AccountID, addrAbbr(distrib.AccountID), distrib.Amount)
				}

				fmt.Fprintf(reportStr, "\n<b>Conflict</b>\n")

				for recommender, recommendeds := range res.Conflict {
					fmt.Fprintf(reportStr, "  <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n", recommender, addrAbbr(recommender))
					for _, recommended := range recommendeds {
						fmt.Fprintf(reportStr, "    <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n", recommended, addrAbbr(recommended))
					}
					fmt.Fprintf(reportStr, "\n")
				}

				fmt.Fprintf(reportStr, "\n<b>XDR</b>\n")
				fmt.Fprintf(reportStr, "<code>%s</code>", xdr)

				msg := tgbotapi.NewMessage(st.UserID, reportStr.String())
				msg.ReplyMarkup = resultButtons
				msg.ParseMode = "HTML"
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},
		})
}

func addrAbbr(addr string) string {
	return fmt.Sprintf("%s...%s", addr[:4], addr[len(addr)-4:])
}
