package common

import (
	"testing"
)

func TestSignature(t *testing.T) {
	payload := map[string]interface{}{
		"a":   100,
		"b":   "Rick",
		"abc": "{}",
	}

	Sign("http://127.0.0.1", "test", payload)

	err := SignCheck("http://127.0.0.1", "test", payload)
	if err != nil {
		t.Fail()
	}
}
