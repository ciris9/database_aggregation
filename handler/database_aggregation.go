package handler

import (
	"context"
	"github.com/xxl-job/xxl-job-executor-go"
	"go.uber.org/zap"
	"permission/constants"
	"permission/service"
)

func DatabaseAggregation(cxt context.Context, param *xxl.RunReq) string {
	if err := service.AggregateData(); err != nil {
		zap.S().Error(err)
		return err.Error()
	}
	return constants.DatabaseAggregationTaskDoneKey
}
