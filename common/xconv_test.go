package common_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/rickone/athena/common"
)

type test1 struct {
	A int64
	B string
	C time.Time `xconv:"Unix"`
	D string    `xconv:"-"`
}

type test2 struct {
	A int64
	B string
	C int64
}

func TestXconv(t *testing.T) {
	t1 := test1{
		A: 100,
		B: "Hello",
		C: time.Unix(10000, 0),
		D: "World",
	}
	t2 := &test2{}
	common.Xconv(&t1, t2)

	if t2.A != t1.A || t2.B != t1.B || t2.C != t1.C.Unix() {
		t.Fail()
	}
}

func TestXconvs(t *testing.T) {
	t1 := []*test1{
		{
			A: 100,
			B: "Hello",
			C: time.Unix(10000, 0),
			D: "World",
		},
		{
			A: 200,
			B: "Hello2",
			C: time.Unix(20000, 0),
			D: "World2",
		},
	}
	t2 := common.Xconv(t1, &test2{}).([]*test2)

	fmt.Printf("%v\n", t2)
	if len(t2) != 2 || t2[0].A != t1[0].A || t2[1].C != t1[1].C.Unix() {
		t.Fail()
	}
}
