package ginex

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// 返回obj.Result, 如果存在
// gRPC的接口返回的结构里有一种情况是查询后的多个结果，必须用一个message包起来
func getResultIfExists(obj interface{}) interface{} {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return obj
	}

	arrVal := val.FieldByName("Result")
	if arrVal.IsValid() {
		return arrVal.Interface()
	}

	return obj
}

func Wrap(f func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		obj, err := f(c)
		if err != nil {
			handleError(c, err)
			return
		}

		if obj != nil {
			switch c.Request.Method {
			case "POST":
				c.JSON(http.StatusCreated, obj)
			case "DELETE":
				c.Status(http.StatusNoContent)
			default:
				//c.SecureJSON(http.StatusOK, obj)
				c.JSON(http.StatusOK, getResultIfExists(obj))
			}
		}
	}
}
