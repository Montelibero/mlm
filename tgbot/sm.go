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

				msg := tgbotapi.NewMessage(st.UserID, makeSummary(
					ctx,
					res.Distributes,
					res.Conflicts,
					res.XDR,
				))
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

				conflicts, err := t.q.GetReportConflicts(ctx, id)
				if err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, makeSummary(ctx, distribs, conflicts, rep.Xdr))
				msg.ReplyMarkup = reportResultButtons
				msg.ParseMode = "HTML"
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},
			eventReportDelete: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				idStr := e.Args[1].(string)
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					e.Cancel(err)
					return
				}

				if err := t.q.DeleteReport(ctx, id); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, "Report deleted")
				msg.ReplyMarkup = reportResultButtons
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
				}
			},
		})
}

func addrAbbr(addr string) string {
	return fmt.Sprintf("%s...%s", addr[:4], addr[len(addr)-4:])
}

func makeSummary(
	_ context.Context,
	distributes []db.ReportDistribute,
	conflicts []db.ReportConflict,
	xdr string,
) string {
	summary := &strings.Builder{}
	fmt.Fprintf(summary, "<b>Report results</b>\n")

	fmt.Fprintf(summary, "\n<b>Distributes</b>\n")

	for _, distrib := range distributes {
		fmt.Fprintf(summary, "<a href=\"https://bsn.mtla.me/accounts/%s\">%s</a> - %f %s\n", distrib.Recommender, addrAbbr(distrib.Recommender), distrib.Amount, distrib.Asset)
	}

	if len(conflicts) > 0 {
		fmt.Fprintf(summary, "\n<b>Conflicts</b>\n")

		for _, conflict := range conflicts {
			fmt.Fprintf(summary, "<a href=\"https://bsn.mtla.me/accounts/%s\">%s</a> - ", conflict.Recommender, addrAbbr(conflict.Recommender))
			fmt.Fprintf(summary, "<a href=\"https://bsn.mtla.me/accounts/%s\">%s</a>\n", conflict.Recommended, addrAbbr(conflict.Recommended))
		}
	}

	fmt.Fprintf(summary, "\n<b>XDR</b>\n")
	fmt.Fprintf(summary, "<code>%s</code>", xdr)

	return summary.String()
}
