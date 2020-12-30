package main

import (
	"context"
	"log"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/eth"
	"github.com/rickone/athena/logger"
)

func main() {
	common.InitRand()
	config.Init("test.yml")
	logger.Init("test")

	trx, err := eth.Client().GetTransaction(context.Background(), "0x4a010f8d0b70eed0b69aff7f817bfc56957774724107792cc6762fb03559d704")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("hash: %s block: %d, from: %s, to: %s, data: %s, %v, fee: %v\n", trx.Hash, trx.GetBlock(), trx.From.Hex(), trx.To.Hex(), trx.GetParamToAddress().Hex(), trx.GetParamValue(), trx.Fee)
}
