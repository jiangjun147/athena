package tron

import (
	"encoding/hex"

	"github.com/rickone/athena/chain"
	"github.com/rickone/athena/errcode"
	"google.golang.org/grpc/status"
)

const (
	emptyAddress = "T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb"
)

func (cli *TronClient) TriggerSmartContract(ownerAddress string, contract string, selector string, parameter chain.Parameter, feeLimit uint64) (*Transaction, error) {
	resp := struct {
		Tx     Transaction `json:"transaction"`
		Result Result      `json:"result"`
	}{}
	err := cli.httpPost("/wallet/triggersmartcontract", map[string]interface{}{
		"owner_address":     ownerAddress,
		"contract_address":  contract,
		"function_selector": selector,
		"parameter":         hex.EncodeToString(parameter),
		"fee_limit":         feeLimit,
		"visible":           true,
	}, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result.Result {
		return nil, status.Errorf(errcode.ErrChainFailed, "TriggerSmartContract err: code=%s msg=%s", resp.Result.Code, resp.Result.Message)
	}
	return &resp.Tx, err
}

func (cli *TronClient) TriggerConstantContract(contract string, selector string, parameter chain.Parameter) (chain.Parameter, error) {
	resp := struct {
		ConstantResult []string `json:"constant_result"`
		Result         Result   `json:"result"`
	}{}
	err := cli.httpPost("/wallet/triggerconstantcontract", map[string]interface{}{
		"owner_address":     emptyAddress,
		"contract_address":  contract,
		"function_selector": selector,
		"parameter":         hex.EncodeToString(parameter),
		"visible":           true,
	}, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result.Result {
		return nil, status.Errorf(errcode.ErrChainFailed, "TriggerConstantContract err: code=%s msg=%s", resp.Result.Code, resp.Result.Message)
	}

	return chain.NewFromHexParameter(resp.ConstantResult)
}
