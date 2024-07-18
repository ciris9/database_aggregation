package service

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"log"
	"permission/config"
	"permission/constants"
	"strconv"
	"strings"
)

func AggregateData() error {
	//通过xxl-job定时拉取，拉取之前需要重置一下从apollo获取的数据
	if err := config.C.GetConfigFromApollo(); err != nil {
		return err
	}
	pullDatabaseData()
	pushDatabaseData()
	return nil
}

// pullDatabase 目前来讲，对于所有的database查询的权限表字段要求格式统一
// 例如：user_id(primary key) name email p (p为标识权限的字段) b1 b2 b3 (对于p，b1 权限有无，b2 权限有无 b3 权限有无)
// 目前暂时先假定位只有可读可写两个权限
func pullDatabaseData() {
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
			zap.S().Info(columns)
			databaseDataCache.Columns(database.DBName, columns)

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
				for _, col := range values {
					if col == nil {
						value = constants.DatabaseNullData
					} else {
						value = string(col)
					}
					rowData = append(rowData, value)
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
// 权限数据库表字段暂定为，user_id(primary key) name email p (p为标识权限的字段) b1 b2 b3 (对于p，b1 权限有无，b2 权限有无 b3 权限有无)
// 聚合到单个数据库，可以暂定为
// user_id(primary key) name email dbname p (p为标识权限的字段) b1 b2 b3 (对于p，b1 权限有无，b2 权限有无 b3 权限有无)
// 这里就假定，只有可读，可写两个权限，表名为pushDatabase.dbName
func pushDatabaseData() {
	for _, pushDatabase := range config.C.TimerMiddlewareConfig.PushDatabaseConfig {
		db, err := sql.Open(constants.MysqlDriverName, pushDatabase.DSN)
		if err != nil {
			log.Fatal(err)
		}

		//todo 定时任务插入数据库之前，需要清空数据库
		{
			stmt, err := db.Prepare(fmt.Sprintf("TRUNCATE TABLE %s", pushDatabase.DBName))
			if err != nil {
				zap.S().Panic(err)
			}
			_, err = stmt.Exec()
			if err != nil {
				zap.S().Panic(err)
			}
			if err := stmt.Close(); err != nil {
				zap.S().Panic(err)
			}
		}

		for dbname, data := range databaseDataCache.Data {
			//todo 插入聚合数据 INSERT INTO employees (first_name, last_name) VALUES ('John', 'Doe');
			insertSql := getInsertSql(pushDatabase.DBName, dbname, data)
			//执行插入数据操作
			{
				stmt, err := db.Prepare(insertSql)
				if err != nil {
					zap.S().Panic(err)
				}
				_, err = stmt.Exec()
				if err != nil {
					zap.S().Panic(err)
				}
				err = stmt.Close()
				if err != nil {
					zap.S().Panic(err)
				}
			}
		}
		if err = db.Close(); err != nil {
			zap.S().Error(err)
		}
	}
	databaseDataCache.Clear()
}

// 拼接sql，插入目标数据库
func getInsertSql(target, dbname string, data *dbFields) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("INSERT INTO %s(", target))
	builder.WriteString("dbname,")
	for i, column := range data.Columns {
		if column == constants.DatabaseColumnID {
			continue
		}
		builder.WriteString(column)
		if i != len(data.Columns)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(") VALUES ")
	zap.S().Info(builder.String())
	for i, row := range data.Rows {
		builder.WriteString("('")
		builder.WriteString(dbname)
		if len(row) != 0 {
			builder.WriteString("',")
		}
		for j, col := range row {
			if col == constants.DatabaseColumnID {
				continue
			}
			if _, err := strconv.ParseInt(col, 10, 64); err == nil {
				builder.WriteString(col)
			} else {
				builder.WriteString("'")
				builder.WriteString(col)
				builder.WriteString("'")
			}
			if j != len(row)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString(")")
		if i != len(data.Rows)-1 {
			builder.WriteString(",")
		} else {
			builder.WriteString(";")
		}
	}
	zap.S().Info("insert sql:", builder.String())
	return builder.String()
}
