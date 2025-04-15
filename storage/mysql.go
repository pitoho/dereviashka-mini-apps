package storage

import (
	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

var (
	DB  *sql.DB

)

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	
	err = DB.Ping()
	if err != nil {
		return err
	}
	
	return nil
}
func GetDB() *sql.DB {
    return DB
}