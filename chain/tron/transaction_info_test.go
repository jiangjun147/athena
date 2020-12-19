package tron_test

import (
	"encoding/json"
	"testing"

	"github.com/rickone/athena/chain/tron"
	"github.com/shopspring/decimal"
)

func TestTransactionInfo(t *testing.T) {
	str := `{"receipt":{"energy_fee":"10000000000001234567890123456789", "net_fee":234.12345699999991}}`
	ti := tron.TransactionInfo{}
	err := json.Unmarshal([]byte(str), &ti)

	if !ti.Receipt.EnergyFee.GreaterThan(decimal.Zero) {
		t.Fail()
	}

	if !ti.Receipt.NetFee.GreaterThan(decimal.Zero) {
		t.Fail()
	}
	t.Logf("ti=%+v err=%v", ti, err)
}
