package distributor

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/Montelibero/mlm"
	"github.com/Montelibero/mlm/stellar"
)

const Account = "GDWXHJJZDQNR5OUGVWEVEXBSBQ6GQSKULOIXFQ63PCOXRSPOOQTYMMLM" // TODO: move to config

var ErrNoBalance = errors.New("no balance")

type Distributor struct {
	stellar mlm.StellarAgregator
}

func (d *Distributor) Distribute(ctx context.Context) (*mlm.DistributeResult, error) {
	// TODO: last start

	// TODO: lock

	distributeAmount, err := d.getDistributeAmount(ctx)
	if err != nil {
		return nil, err
	}

	recs, err := d.stellar.Recommenders(ctx)
	if err != nil {
		return nil, err
	}

	res := &mlm.DistributeResult{
		Conflict: recs.Conflict,
	}
	part := distributeAmount / float64(recs.TotalRecommendedMTLAP)

	for _, recommender := range recs.Recommenders {
		partCount := 0

		for _, recommended := range recommender.Recommended {
			if _, ok := recs.Conflict[recommended.AccountID]; ok {
				continue
			}

			partCount += recommended.MTLAPCount
		}

		res.Distributes = append(res.Distributes, mlm.Distribute{
			AccountID: recommender.AccountID,
			Amount:    math.Floor(float64(partCount)*part*10000000) / 10000000,
		})
	}

	return res, nil
}

func (d *Distributor) getDistributeAmount(ctx context.Context) (float64, error) {
	balstr, err := d.stellar.Balance(ctx, Account, stellar.EURMTLAsset, stellar.EURMTLIssuer)
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

	return bal, nil
}

func New(stellar mlm.StellarAgregator) *Distributor {
	return &Distributor{stellar: stellar}
}

var _ mlm.Distributor = &Distributor{}
