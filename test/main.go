package main

import (
	"context"
	"log"

	"github.com/rickone/athena/chain/tron"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/logger"
)

func main() {
	common.InitRand()
	config.Init("test.yml")
	logger.Init("test")

	trx, err := tron.Client().GetTransaction(context.Background(), "e4be5e35b692d0926520371951313a6b34da9571e258b90167a449ff540f5f21")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("trx: %+v\n", trx)
	if trx.Data != nil {
		log.Printf("data: %+v\n", *trx.Data)
	}
}
