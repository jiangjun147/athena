package ginex

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/rickone/athena/errcode"
)

type HandlerInvoker interface {
	Invoke(c *gin.Context) (interface{}, error)
}

func Bind(val HandlerInvoker) gin.HandlerFunc {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return Wrap(func(c *gin.Context) (interface{}, error) {
		v := reflect.New(t)
		if err := c.ShouldBind(v.Interface()); err != nil {
			return nil, Error(errcode.ErrGinBind, err.Error())
		}

		invoker := v.Interface().(HandlerInvoker)
		return invoker.Invoke(c)
	})
}
