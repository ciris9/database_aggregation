package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"testing"
)

func TestViper(t *testing.T) {
	conf := &config{viper: viper.New()}
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath(".")
	if err := conf.viper.ReadInConfig(); err != nil {
		zap.S().Error(err)
	}
	zap.S().Info(conf.viper.AllKeys())
}
