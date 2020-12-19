package common

import (
	"math/rand"
	"time"
)

const (
	runes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ23456789"
)

func InitRand() {
	const mark int64 = ((1 << 32) - 1)
	ns := time.Now().UnixNano()
	high := (ns >> 32) & mark
	low := ns & mark
	rand.Seed((low << 32) | high)
}

func RandomString(n int) string {
	return RandomStringFill(n, runes)
}

func RandomStringFill(n int, fillChars string) string {
	var code = ""
	for i := 0; i < n; i++ {
		code += string(fillChars[rand.Intn(len(fillChars))])
	}
	return code
}
