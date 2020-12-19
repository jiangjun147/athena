package ginex

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rickone/athena/errcode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(code int, msg string) error {
	return status.Error(codes.Code(code), msg)
}

func handleError(c *gin.Context, err error) {
	err = errcode.ErrorMap(err)

	st, ok := status.FromError(err)
	if ok {
		handleCodeMsg(c, int(st.Code()), st.Message())
		return
	}

	handleCodeMsg(c, errcode.ErrRpcFailed, err.Error())
}

func handleCodeMsg(c *gin.Context, code int, msg string) {
	statusCode := http.StatusInternalServerError
	if code > 1000 {
		statusCode = code / 1000
	}

	c.JSON(statusCode, gin.H{
		"Code": code,
		"Msg":  msg,
	})
}
