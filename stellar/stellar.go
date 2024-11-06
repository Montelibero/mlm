package stellar

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/Montelibero/mlm"
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
)

const (
	DefaultLimit      = 20
	MTLAPAsset        = "MTLAP"
	MTLAPIssuer       = "GCNVDZIHGX473FEI7IXCUAEXUJ4BGCKEMHF36VYP5EMS7PX2QBLAMTLA"
	MTLAPAssetRequest = "MTLAP:GCNVDZIHGX473FEI7IXCUAEXUJ4BGCKEMHF36VYP5EMS7PX2QBLAMTLA"
	TagRecommend      = "RecommendToMTLA"
)

type Client struct {
	cl mlm.HorizonClient
}

func (c *Client) Fetch(ctx context.Context) (*mlm.RecommenderFetchResult, error) {
	var allAccounts []horizon.Account
	accp, err := c.cl.Accounts(horizonclient.AccountsRequest{
		Asset: MTLAPAssetRequest,
		Limit: DefaultLimit,
	})
	if err != nil {
		return nil, err
	}
	if len(accp.Embedded.Records) < DefaultLimit {
		return accountsToResult(accp.Embedded.Records), nil
	}

	for {
		allAccounts = append(allAccounts, accp.Embedded.Records...)
		accp, err = c.cl.NextAccountsPage(accp)
		if err != nil {
			return nil, err
		}
		if len(accp.Embedded.Records) == 0 {
			break
		}
	}

	return accountsToResult(allAccounts), nil
}

func NewClient(cl horizonclient.ClientInterface) *Client {
	return &Client{cl: cl}
}

func accountsToResult(accs []horizon.Account) *mlm.RecommenderFetchResult {
	res := &mlm.RecommenderFetchResult{}

	accMap := lo.Associate(accs, func(acc horizon.Account) (string, horizon.Account) {
		return acc.AccountID, acc
	})

	for _, recommender := range accs {
		recommendedDataMap := lo.PickBy(recommender.Data, func(k, v string) bool {
			return strings.HasPrefix(k, TagRecommend)
		})

		if len(recommendedDataMap) == 0 {
			continue
		}

		recommendeds := make([]mlm.Recommended, 0, len(recommendedDataMap))
		for _, v := range recommendedDataMap {
			accountIDRaw, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				spew.Dump(err)
			}
			accountID := string(accountIDRaw)
			recommended, ok := accMap[accountID]
			if !ok {
				continue
			}

			mtlapBalance, _ := strconv.ParseFloat(recommended.GetCreditBalance(MTLAPAsset, MTLAPIssuer), 64)
			mtlapCount := int(mtlapBalance)

			res.TotalRecommendedMTLAP += mtlapCount

			recommendeds = append(recommendeds, mlm.Recommended{
				AccountID:  recommended.AccountID,
				MTLAPCount: mtlapCount,
			})
		}

		res.Recommenders = append(res.Recommenders, mlm.Recommender{
			AccountID:   recommender.AccountID,
			Recommended: recommendeds,
		})
	}

	return res
}

var _ mlm.RecommenderFetcher = &Client{}
