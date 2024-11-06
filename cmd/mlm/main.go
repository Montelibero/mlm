package main

import (
	"context"
	"fmt"

	"github.com/Montelibero/mlm"
	"github.com/Montelibero/mlm/distributor"
	"github.com/Montelibero/mlm/stellar"
	"github.com/samber/lo"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
)

func main() {
	ctx := context.Background()

	cl := horizonclient.DefaultPublicNetClient

	stell := stellar.NewClient(cl)
	distrib := distributor.New(stell)

	res, err := distrib.Distribute(ctx)
	if err != nil {
		panic(err)
	}

	accountDetail, err := cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: distributor.Account,
	})
	if err != nil {
		panic(err)
	}

	ops := lo.Map(res.Distributes, func(d mlm.Distribute, _ int) txnbuild.Operation {
		return &txnbuild.Payment{
			Destination: d.AccountID,
			Amount:      fmt.Sprintf("%.7f", d.Amount),
			Asset:       txnbuild.CreditAsset{Code: stellar.EURMTLAsset, Issuer: stellar.EURMTLIssuer},
		}
	})

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accountDetail,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              10000,
		Memo:                 txnbuild.MemoText("mtla mlm distribution"),
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	})
	if err != nil {
		panic(err)
	}

	xdr, err := tx.Base64()
	if err != nil {
		panic(err)
	}

	fmt.Println(xdr)
}
