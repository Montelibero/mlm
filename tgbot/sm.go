package tgbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Montelibero/mlm/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
)

const (
	// states
	stateInit         = "state_init"
	stateReports      = "state_reports"
	stateReportResult = "state_report_result"

	// events
	eventStart        = "/start"
	eventReportRun    = "/report_run"
	eventReportDelete = "/report_delete"
	eventReportResult = "/report_result"
)

var startButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(eventStart),
		tgbotapi.NewKeyboardButton(eventReportRun),
	),
)

var reportResultButtons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(eventStart),
	),
)

func (t *TGBot) newSM() *fsm.FSM {
	return fsm.NewFSM(
		stateInit,
		fsm.Events{
			{Name: eventStart, Src: []string{stateInit, stateReports, stateReportResult}, Dst: stateReports},
			{Name: eventReportRun, Src: []string{stateReports}, Dst: stateReportResult},
			{Name: eventReportResult, Src: []string{stateReports}, Dst: stateReportResult},
			{Name: eventReportDelete, Src: []string{stateReports, stateReportResult}, Dst: stateReports},
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				if err := t.q.CreateState(ctx, db.CreateStateParams{
					UserID: st.UserID,
					State:  e.Dst,
					Data:   st.Data,
					Meta:   st.Meta,
				}); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventStart: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				rr, err := t.q.GetReports(ctx, 12)
				if err != nil {
					e.Cancel(err)
					return
				}

				scr := &strings.Builder{}

				fmt.Fprintf(scr, "<b>Reports</b>\n")

				for _, r := range rr {
					fmt.Fprintf(scr, "  /report_result_%d %s\n", r.ID, r.CreatedAt.Time.Format(time.DateOnly))
				}

				if len(rr) == 0 {
					fmt.Fprintf(scr, "  <i>nothing to report</i>\n")
				}

				msg := tgbotapi.NewMessage(st.UserID, scr.String())
				msg.ReplyMarkup = startButtons
				msg.ParseMode = "HTML"
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},

			eventReportRun: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				if _, err := t.bot.Send(tgbotapi.NewMessage(st.UserID, "Processing...")); err != nil {
					e.Cancel(err)
					return
				}

				res, err := t.distrib.Distribute(ctx)
				if err != nil {
					e.Cancel(err)
					return
				}

				summary := &strings.Builder{}
				fmt.Fprintf(summary, "<b>Report results</b>\n")

				for _, distrib := range res.Distributes {
					fmt.Fprintf(summary, "  <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a> - %f EURMTL\n", distrib.AccountID, addrAbbr(distrib.AccountID), distrib.Amount)
				}

				fmt.Fprintf(summary, "\n<b>Conflict</b>\n")

				for recommender, recommendeds := range res.Conflict {
					fmt.Fprintf(summary, "  <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n", recommender, addrAbbr(recommender))
					for _, recommended := range recommendeds {
						fmt.Fprintf(summary, "    <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n", recommended, addrAbbr(recommended))
					}
					fmt.Fprintf(summary, "\n")
				}

				fmt.Fprintf(summary, "\n<b>XDR</b>\n")
				fmt.Fprintf(summary, "<code>%s</code>", res.XDR)

				msg := tgbotapi.NewMessage(st.UserID, summary.String())
				msg.ReplyMarkup = reportResultButtons
				msg.ParseMode = "HTML"
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},
			eventReportResult: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				idStr := e.Args[1].(string)
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					e.Cancel(err)
					return
				}

				rep, err := t.q.GetReport(ctx, id)
				if err != nil {
					e.Cancel(err)
					return
				}

				distribs, err := t.q.GetReportDistributes(ctx, id)
				if err != nil {
					e.Cancel(err)
					return
				}

				// TODO: make one report message
				summary := &strings.Builder{}
				fmt.Fprintf(summary, "<b>Report #%d %s</b>\n", rep.ID, rep.CreatedAt.Time.Format(time.DateOnly))

				for _, distrib := range distribs {
					fmt.Fprintf(summary, "  <a href=\"https://bsn.mtla.me/accounts/%s\">%s</a> - %f %s\n", distrib.Recommender, addrAbbr(distrib.Recommender), distrib.Amount, distrib.Asset)
				}

				fmt.Fprintf(summary, "\n<b>XDR</b>\n")
				fmt.Fprintf(summary, "<code>%s</code>", rep.Xdr)

				msg := tgbotapi.NewMessage(st.UserID, summary.String())
				msg.ReplyMarkup = reportResultButtons
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
