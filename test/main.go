package main

import (
	"log"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/logger"
)

func main() {
	common.InitRand()
	config.Init("test.yml")
	logger.Init("test")

	ids := []int64{8192, 8193, 80941583532363776, 80941583532363777}
	for _, id := range ids {
		addr := common.EncodeAddress(id, 1)
		log.Printf("addr=%s\n", addr)

		id, chainId, err := common.DecodeAddress(addr)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("id: %d idr: %d chain-id: %d\n", id, ^id, chainId)
	}
}
