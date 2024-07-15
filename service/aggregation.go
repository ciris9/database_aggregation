package service

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"log"
	"permission/config"
	"permission/constants"
)

func AggregateData() error {
	//通过xxl-job定时拉取，拉取之前需要重置一下从apollo获取的数据
	if err := config.C.GetConfigFromApollo(); err != nil {
		return err
	}
	pullDatabase()
	pushDatabase()
	return nil
}

// pullDatabase 目前来讲，对于所有的database查询的权限表字段要求格式统一
// 例如：id(primary key) name email p1 p2 p3 p4 (p为标识权限的字段，例如为1则有权限，为0则无权限)
func pullDatabase() {
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

// pushDatabase 推送到目标数据库的表字段基本是固定的
func pushDatabase() {
	for _, database := range config.C.TimerMiddlewareConfig.PushDatabaseConfig {
		db, err := sql.Open(constants.MysqlDriverName, database.DSN)
		if err != nil {
			log.Fatal(err)
		}
		for _, sqlSentence := range database.Sqls {
			exec, err := db.Exec(sqlSentence)
			if err != nil {
				zap.S().Error(err)
			}
			affected, err := exec.RowsAffected()
			if err != nil {
				zap.S().Error(err)
			}
			id, err := exec.LastInsertId()
			if err != nil {
				zap.S().Error(err)
			}
			zap.S().Infof("push database affected:%d id: %d", affected, id)
		}
		if err = db.Close(); err != nil {
			zap.S().Error(err)
		}
	}
	databaseDataCache.Clear()
}
