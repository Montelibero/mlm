package distributor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Montelibero/mlm"
	"github.com/Montelibero/mlm/config"
	"github.com/Montelibero/mlm/db"
	"github.com/Montelibero/mlm/stellar"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/stellar/go/txnbuild"
)

var ErrNoBalance = errors.New("no balance")

type Distributor struct {
	cfg     *config.Config
	stellar mlm.StellarAgregator
	q       *db.Queries
	pg      *pgx.Conn
}

func (d *Distributor) Distribute(ctx context.Context) (*mlm.DistributeResult, error) {
	if err := d.q.LockReport(ctx); err != nil {
		return nil, err
	}
	defer func() { _ = d.q.UnlockReport(ctx) }()

	lastDistribute, err := d.getLastDistribute(ctx)
	if err != nil {
		return nil, err
	}

	distributeAmount, err := d.getDistributeAmount(ctx)
	if err != nil {
		return nil, err
	}

	recs, err := d.stellar.Recommenders(ctx)
	if err != nil {
		return nil, err
	}

	res, err := d.calcuateParts(ctx, lastDistribute, distributeAmount, recs)
	if err != nil {
		return nil, err
	}

	res.XDR, err = d.getXDR(ctx, res.Distributes)
	if err != nil {
		return nil, err
	}

	res.ReportID, err = d.createReport(ctx, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d *Distributor) getLastDistribute(ctx context.Context) (map[string]map[string]int64, error) {
	rr, err := d.q.GetReports(ctx, 1)
	if err != nil {
		return nil, err
	}

	lastDistribute := map[string]map[string]int64{} // recommender-recommended-mtlap

	if len(rr) > 0 {
		ras, err := d.q.GetReportRecommends(ctx, rr[0].ID)
		if err != nil {
			return nil, err
		}

		for _, ra := range ras {
			if _, ok := lastDistribute[ra.Recommender]; !ok {
				lastDistribute[ra.Recommender] = make(map[string]int64)
			}

			lastDistribute[ra.Recommender][ra.Recommended] = ra.RecommendedMtlap
		}
	}

	return lastDistribute, nil
}

func (d *Distributor) getDistributeAmount(ctx context.Context) (float64, error) {
	balstr, err := d.stellar.Balance(ctx, d.cfg.Address, stellar.EURMTLAsset, stellar.EURMTLIssuer)
	if err != nil {
		return 0, err
	}

	bal, err := strconv.ParseFloat(balstr, 64)
	if err != nil {
		return 0, err
	}

	if bal == 0 {
		return 0, ErrNoBalance
	}

	return bal / 3 * 10000000 / 10000000, nil
}

func (d *Distributor) calcuateParts(
	ctx context.Context,
	lastDistribute map[string]map[string]int64,
	distributeAmount float64,
	recs *mlm.RecommendersFetchResult,
) (*mlm.DistributeResult, error) {
	res := &mlm.DistributeResult{
		Conflicts:   make([]db.ReportConflict, 0),
		Recommends:  make([]db.ReportRecommend, 0),
		Distributes: make([]db.ReportDistribute, 0),
	}
	part := distributeAmount / float64(recs.TotalRecommendedMTLAP)

	for recommended, recommenders := range recs.Conflict {
		for _, recoomender := range recommenders {
			res.Conflicts = append(res.Conflicts, db.ReportConflict{
				Recommender: recoomender,
				Recommended: recommended,
			})
		}
	}

	for _, recommender := range recs.Recommenders {
		partCount := int64(0)

		for _, recommended := range recommender.Recommended {
			if _, ok := recs.Conflict[recommended.AccountID]; ok {
				continue
			}

			lastMTLAP, _ := lastDistribute[recommender.AccountID][recommended.AccountID]
			if lastMTLAP < recommended.MTLAP { // dynamics
				partCount += recommended.MTLAP - lastMTLAP
			}

			res.Recommends = append(res.Recommends, db.ReportRecommend{
				Recommender:      recommender.AccountID,
				Recommended:      recommended.AccountID,
				RecommendedMtlap: recommended.MTLAP,
			})
		}

		res.Distributes = append(res.Distributes, db.ReportDistribute{
			Recommender: recommender.AccountID,
			Asset:       stellar.EURMTLAsset,
			Amount:      math.Floor(float64(partCount)*part*10000000) / 10000000,
		})
	}

	return res, nil
}

func (d *Distributor) getXDR(ctx context.Context, distributes []db.ReportDistribute) (string, error) {
	accountDetail, err := d.stellar.AccountDetail(d.cfg.Address)
	if err != nil {
		return "", err
	}

	ops := lo.Map(distributes, func(d db.ReportDistribute, _ int) txnbuild.Operation {
		return &txnbuild.Payment{
			Destination: d.Recommender,
			Amount:      fmt.Sprintf("%.7f", d.Amount),
			Asset:       txnbuild.CreditAsset{Code: d.Asset, Issuer: stellar.EURMTLIssuer},
		}
	})

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accountDetail,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              1000,
		Memo:                 txnbuild.MemoText(fmt.Sprintf("mlta mlm %s", time.Now().Format(time.DateOnly))),
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	})
	if err != nil {
		return "", err
	}

	xdr, err := tx.Base64()
	if err != nil {
		return "", err
	}

	return xdr, err
}

func (d *Distributor) createReport(ctx context.Context, res *mlm.DistributeResult) (int64, error) {
	tx, err := d.pg.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	qtx := d.q.WithTx(tx)

	reportID, err := qtx.CreateReport(ctx, res.XDR)
	if err != nil {
		return 0, err
	}

	for _, recommend := range res.Recommends {
		if err := qtx.CreateReportRecommend(ctx, db.CreateReportRecommendParams{
			ReportID:         reportID,
			Recommender:      recommend.Recommender,
			Recommended:      recommend.Recommended,
			RecommendedMtlap: recommend.RecommendedMtlap,
		}); err != nil {
			return 0, err
		}
	}

	for _, distrib := range res.Distributes {
		if err := qtx.CreateReportDistribute(ctx, db.CreateReportDistributeParams{
			ReportID:    reportID,
			Recommender: distrib.Recommender,
			Asset:       distrib.Asset,
			Amount:      distrib.Amount,
		}); err != nil {
			return 0, err
		}
	}

	for _, conflict := range res.Conflicts {
		if err := qtx.CreateReportConflict(ctx, db.CreateReportConflictParams{
			ReportID:    reportID,
			Recommender: conflict.Recommender,
			Recommended: conflict.Recommended,
		}); err != nil {
			return 0, err
		}
	}

	return reportID, tx.Commit(ctx)
}

func New(
	cfg *config.Config,
	stellar mlm.StellarAgregator,
	q *db.Queries,
	pg *pgx.Conn,
) *Distributor {
	return &Distributor{
		cfg:     cfg,
		stellar: stellar,
		q:       q,
		pg:      pg,
	}
}

var _ mlm.Distributor = &Distributor{}
