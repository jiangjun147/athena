package errcode

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/status"
)

func ErrorMap(err error) error {
	if gorm.IsRecordNotFoundError(err) {
		return status.Error(ErrRecordNotFound, "record not found")
	}

	switch err {
	case redis.ErrNil:
		err = status.Error(ErrValueNotFound, "value not found")
	}
	return err
}
