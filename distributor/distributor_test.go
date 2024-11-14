package distributor_test

import (
	"context"
	"testing"

	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	"github.com/davecgh/go-spew/spew"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stretchr/testify/require"
)

func TestDistributor_Distribute(t *testing.T) {
	t.Skip(t) // TODO: implement

	ctx := context.Background()
	recs := stellar.NewClient(horizonclient.DefaultPublicNetClient)
	distr := distributor.New(recs, nil, nil)

	res, err := distr.Distribute(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	spew.Dump(res)
}
