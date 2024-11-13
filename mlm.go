package mlm

import (
	"context"

	"github.com/stellar/go/clients/horizonclient"
)

type Recommended struct {
	AccountID  string
	MTLAPCount int
}

type Recommender struct {
	AccountID   string
	Recommended []Recommended
}

type RecommendersFetchResult struct {
	Conflict              map[string][]string // recommended-recommender
	Recommenders          []Recommender
	TotalRecommendedMTLAP int
}

type StellarAgregator interface {
	Balance(ctx context.Context, accountID, asset, issuer string) (string, error)
	Recommenders(ctx context.Context) (*RecommendersFetchResult, error)
}

type HorizonClient interface {
	horizonclient.ClientInterface
}

type Distribute struct {
	AccountID string
	Amount    float64
}

type DistributeResult struct {
	Conflict    map[string][]string // recommended-recommender
	Distributes []Distribute
}

type Distributor interface {
	Distribute(ctx context.Context) (*DistributeResult, error)
}
