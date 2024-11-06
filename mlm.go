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

type RecommenderFetchResult struct {
	Recommenders          []Recommender
	TotalRecommendedMTLAP int
}

type RecommenderFetcher interface {
	Fetch(ctx context.Context) (*RecommenderFetchResult, error)
}

type HorizonClient interface {
	horizonclient.ClientInterface
}

type DistributeResult struct{}

type Distributor interface {
	Distribute(ctx context.Context) (*DistributeResult, error)
}
