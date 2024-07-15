package service

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"log"
	"permission/config"
	"permission/constants"
	"testing"
)

func TestDatabaseAggregation(t *testing.T) {
	for _, database := range config.C.TimerMiddlewareConfig.PullDatabaseConfig {
		db, err := sql.Open(constants.MysqlDriverName, database.DSN)
		if err != nil {
			log.Fatal(err)
		}
		databaseDataCache.Data[database.DBName] = &dbFields{
			Columns: make([]string, 0),
			Rows:    make([][]string, 0),
		}
		for _, sqlSentence := range database.Sqls {
			//todo for example /  select * from permission1 as p1
			rows, err := db.Query(sqlSentence)
			if err != nil {
				zap.S().Error(err)
			}

			columns, err := rows.Columns()
			if err != nil {
				zap.S().Error(err)
			}
			databaseDataCache.OnceColumns(database.DBName, columns)

			values := make([]sql.RawBytes, len(columns))
			scanArgs := make([]interface{}, len(values))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			for rows.Next() {
				err = rows.Scan(scanArgs...)
				if err != nil {
					log.Fatal(err)
				}
				var value string
				rowData := make([]string, 0)
				for i, col := range values {
					if col == nil {
						value = "NULL"
					} else {
						value = string(col)
					}
					rowData = append(rowData, value)
					fmt.Println(columns[i], ": ", value)
				}
				databaseDataCache.MergeSlice(database.DBName, rowData)
			}
			if err = rows.Err(); err != nil {
				zap.S().Error(err)
			}
		}
		if err = db.Close(); err != nil {
			zap.S().Error(err)
		}
	}
}
