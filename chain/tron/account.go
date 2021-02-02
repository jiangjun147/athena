package tron

import "math/big"

type Account struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
}

func (cli *TronClient) GetAccount(b58Address string) (*Account, error) {
	acc := Account{}
	err := cli.httpPost("/wallet/getaccount", map[string]interface{}{
		"address": b58Address,
		"visible": "true",
	}, &acc)
	return &acc, err
}
