package server

import (
	"github.com/xxl-job/xxl-job-executor-go"
	"permission/config"
	"permission/logs"
	"strconv"
)

type XXLExecutorServer struct {
	Executor xxl.Executor
}

func NewXXLExecutorServer() *XXLExecutorServer {
	return &XXLExecutorServer{}
}

func (s *XXLExecutorServer) startXXLJobClient() {
	s.Executor = xxl.NewExecutor(
		xxl.RegistryKey(config.C.XXLJobConfig.AppName), //执行器名称
		xxl.ServerAddr(config.C.XXLJobConfig.AdminAddress),
		xxl.AccessToken(config.C.XXLJobConfig.Token),                     //请求令牌(默认为空)
		xxl.ExecutorPort(strconv.Itoa(config.C.XXLJobConfig.ClientPort)), //默认9999（非必填）
		xxl.SetLogger(&logs.Logger{}),                                    //自定义日志
		xxl.ExecutorIp(config.C.XXLJobConfig.ExecutorIp),                 //可自动获取
	)
	s.Executor.Init()
	//设置日志查看handler
	s.Executor.LogHandler(logs.CustomLogHandle)
	InitTask(s.Executor)
}

func (s *XXLExecutorServer) Run() error {
	s.startXXLJobClient()
	return s.Executor.Run()
}
