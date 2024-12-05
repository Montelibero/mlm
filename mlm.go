package mlm

import (
	"context"
	"embed"

	"github.com/Montelibero/mlm/db"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

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

type DistributeResult struct {
	XDR         string
	Conflicts   []db.ReportConflict
	Recommends  []db.ReportRecommend
	Distributes []db.ReportDistribute
}

type Distributor interface {
	Distribute(ctx context.Context) (*DistributeResult, error)
}
