package common

import "testing"

func Assert(condition bool, info string) {
	if !condition {
		panic(info)
	}
}

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}

func AssertT(t *testing.T, condition bool) {
	if !condition {
		t.FailNow()
	}
}

func AssertErrorT(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
