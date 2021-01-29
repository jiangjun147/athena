package common

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func Sign(url string, key string, payload map[string]interface{}) {
	now := time.Now()
	payload["timestamp"] = now.Unix()

	strs := []string{url, key}
	for _, v := range payload {
		strs = append(strs, fmt.Sprintf("%v", v))
	}

	sign := Sha256Sign(strs...)
	payload["sign"] = sign
}

func SignCheck(url string, key string, payload map[string]interface{}) error {
	timestamp, ok := payload["timestamp"]
	if !ok {
		return errors.New("no timestamp")
	}

	if math.Abs(float64(time.Now().Unix()-timestamp.(int64))) > 300 {
		return errors.New("sign expired")
	}

	sign := payload["sign"].(string)
	delete(payload, "sign")

	strs := []string{url, key}
	for _, v := range payload {
		strs = append(strs, fmt.Sprintf("%v", v))
	}

	if sign != Sha256Sign(strs...) {
		return errors.New("sign dismatch")
	}
	return nil
}
