package mlm

import (
	"context"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
)

type Recommended struct {
	AccountID string
	MTLAP     int64
}

type Recommender struct {
	AccountID   string
	Recommended []Recommended
}

type RecommendersFetchResult struct {
	Conflict              map[string][]string // recommended-recommender
	Recommenders          []Recommender
	TotalRecommendedMTLAP int64
}

type StellarAgregator interface {
	Balance(ctx context.Context, accountID, asset, issuer string) (string, error)
	Recommenders(ctx context.Context) (*RecommendersFetchResult, error)
	AccountDetail(accountID string) (horizon.Account, error)
}

type HorizonClient interface {
	horizonclient.ClientInterface
}

type Recommend struct {
	Recommender      string
	Recommended      string
	RecommendedMTLAP int64
}

type Distribute struct {
	AccountID string
	Amount    float64
}

type DistributeResult struct {
	Conflict    map[string][]string // recommended-recommender
	XDR         string
	Recommends  []Recommend
	Distributes []Distribute
}

type Distributor interface {
	Distribute(ctx context.Context) (*DistributeResult, error)
}
