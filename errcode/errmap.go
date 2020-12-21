package errcode

import (
	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/status"
)

func ErrorMap(err error) error {
	if gorm.IsRecordNotFoundError(err) {
		return status.Error(ErrRecordNotFound, "record not found")
	}

	if myErr, ok := err.(*mysql.MySQLError); ok && myErr.Number == 1062 {
		return status.Error(ErrKeyDuplicated, "key duplicated")
	}

	switch err {
	case redis.ErrNil:
		err = status.Error(ErrValueNotFound, "value not found")
	}
	return err
}
