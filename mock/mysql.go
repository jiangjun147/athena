package mock

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/mysql"
)

func InitMysqlCli() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./mock.db")
	common.AssertError(err)

	//db = db.Debug()
	db.LogMode(false)

	mysql.SetDB("write", db)
	mysql.SetDB("read", db)
	return db
}

func InitModel(models ...interface{}) {
	db := mysql.DB("write")

	db.DropTableIfExists(models...)
	db.CreateTable(models...)
}
