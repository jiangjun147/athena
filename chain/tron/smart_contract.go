package tron

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/jsonex"
	"google.golang.org/grpc/status"
)

const (
	emptyAddress = "T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb"
)

func TriggerSmartContract(apiUrl string, ownerAddress string, contract string, selector string, parameter Parameter, feeLimit uint64) (*Transaction, error) {
	// https://cn.developers.tron.network/reference#%E8%A7%A6%E5%8F%91%E6%99%BA%E8%83%BD%E5%90%88%E7%BA%A6

	data, err := json.Marshal(map[string]interface{}{
		"owner_address":     ownerAddress,
		"contract_address":  contract,
		"function_selector": selector,
		"parameter":         hex.EncodeToString(parameter),
		"fee_limit":         feeLimit,
		"visible":           true,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/triggersmartcontract", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	resp := struct {
		Tx     Transaction `json:"transaction"`
		Result Result      `json:"result"`
	}{}
	err = jsonex.UnmarshalFromReader(r.Body, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result.Result {
		return nil, status.Errorf(errcode.ErrBlockChain, "TriggerSmartContract err: code=%s msg=%s", resp.Result.Code, resp.Result.Message)
	}
	return &resp.Tx, err
}

func TriggerConstantContract(apiUrl string, contract string, selector string, parameter Parameter) (Parameter, error) {
	// https://cn.developers.tron.network/reference#triggerconstantcontract

	data, err := json.Marshal(map[string]interface{}{
		"owner_address":     emptyAddress,
		"contract_address":  contract,
		"function_selector": selector,
		"parameter":         hex.EncodeToString(parameter),
		"visible":           true,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/triggerconstantcontract", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	resp := struct {
		ConstantResult []string `json:"constant_result"`
		Result         Result   `json:"result"`
	}{}

	err = jsonex.UnmarshalFromReader(r.Body, &resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result.Result {
		return nil, status.Errorf(errcode.ErrBlockChain, "TriggerConstantContract err: code=%s msg=%s", resp.Result.Code, resp.Result.Message)
	}

	return NewFromHexParameter(resp.ConstantResult)
}
