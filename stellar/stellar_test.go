package stellar_test

import (
	"context"
	"testing"

	"github.com/Montelibero/mlm/stellar"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stretchr/testify/require"
)

func TestClient_Fetch(t *testing.T) {
	ctx := context.Background()
	cl := stellar.NewClient(horizonclient.DefaultPublicNetClient) // TODO: use mock

	res, err := cl.Fetch(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.NotEmpty(t, res.Recommenders)
	require.NotEmpty(t, res.TotalRecommendedMTLAP)
}
