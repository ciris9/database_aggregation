package main

import (
	"go.uber.org/zap"
	"permission/server"
)

func main() {
	if err := server.NewXXLExecutorServer().Run(); err != nil {
		zap.S().Panic(err)
	}
}
