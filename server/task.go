package server

import (
	"github.com/xxl-job/xxl-job-executor-go"
	"permission/handler"
)

func InitTask(exec xxl.Executor) {
	exec.RegTask("database-aggregation", handler.DatabaseAggregation)
}
