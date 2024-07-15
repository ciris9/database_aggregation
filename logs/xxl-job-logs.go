package logs

import (
	"github.com/xxl-job/xxl-job-executor-go"
	"go.uber.org/zap"
)

func CustomLogHandle(req *xxl.LogReq) *xxl.LogRes {
	return &xxl.LogRes{Code: xxl.SuccessCode, Msg: "", Content: xxl.LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   2,
		LogContent:  "这个是自定义日志handler",
		IsEnd:       true,
	}}
}

type Logger struct{}

func (l *Logger) Info(format string, a ...interface{}) {
	zap.S().Infof(format, a...)
}

func (l *Logger) Error(format string, a ...interface{}) {
	zap.S().Error(format, a)
}
