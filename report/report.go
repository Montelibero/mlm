package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/Montelibero/mlm"
)

const bsnViewerPrefix = "https://bsn.mtla.me/accounts/"

func FromDistributeResult(res mlm.DistributeResult) string {
	rep := &strings.Builder{}

	if res.XDR != "" {
		fmt.Fprintf(rep, "<b>Отчет наград за продвижение участников</b>")
	}
	if res.XDR == "" {
		fmt.Fprintf(rep, "<b>Предварительный отчет наград за продвижение участников</b>")
	}

	fmt.Fprintf(rep, `

Дата: %s
Сумма: %f EURMTL
Рекомендателей: %d
Рекомендаций: %d
Новые участники: %d
Участники с повышением уровня: %d
Выплата за тег: %f EURMTL`,
		res.CreatedAt.Format(time.DateOnly),
		res.Amount,
		len(res.Distributes),
		len(res.Recommends),
		res.RecommendedNewCount,
		res.RecommendedLevelUpCount,
		res.AmountPerTag)

	if len(res.Conflicts) > 0 {
		fmt.Fprintf(rep, "\n\n<b>Конфликты</b>\n")

		for _, c := range res.Conflicts {
			fmt.Fprintf(rep, "\n<a href=\"%s\">%s</a> -> <a href=\n%s\">%s</a>",
				strings.Join([]string{bsnViewerPrefix, c.Recommender}, ""),
				accountAbbr(c.Recommender),
				strings.Join([]string{bsnViewerPrefix, c.Recommended}, ""),
				accountAbbr(c.Recommended))
		}
	}

	if len(res.Distributes) > 0 {
		fmt.Fprintf(rep, "\n\n<b>Распределение</b>\n")

		isEmpty := true

		for _, d := range res.Distributes {
			if d.Amount == 0 {
				continue
			}

			isEmpty = false

			fmt.Fprintf(rep, "\n<a href=\"%s\">%s</a>: %.2f",
				strings.Join([]string{bsnViewerPrefix, d.Recommender}, ""),
				accountAbbr(d.Recommender),
				d.Amount)
		}

		if isEmpty {
			fmt.Fprintf(rep, "\nНикто никаких наград не заслужил :(")
		}
	}

	if res.XDR != "" {
		fmt.Fprintf(rep, "\n\n<b>XDR</b>\n<code>\n%s\n</code>", res.XDR)
	}

	return rep.String()
}

func accountAbbr(accountID string) string {
	return accountID[:5] + "..." + accountID[len(accountID)-5:]
}
