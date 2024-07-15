package config

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	agollo_config "github.com/apolloconfig/agollo/v4/env/config"
	"go.uber.org/zap"
	"permission/constants"
	"time"
)

// todo 需要定时从apollo拉取数据，更改拉取数据库方式以及聚合方案

var apolloClient agollo.Client

func init() {
	initApolloClient()
	if err := C.GetConfigFromApollo(); err != nil {
		zap.S().Panic(err)
	}
}

type apolloConfig struct {
	AppID          string `yaml:"AppID" json:"AppID"`
	Cluster        string `yaml:"Cluster" json:"Cluster"`
	NamespaceName  string `yaml:"NamespaceName" json:"NamespaceName"`
	Address        string `yaml:"Address" json:"Address"`
	IsBackupConfig bool   `yaml:"IsBackupConfig" json:"IsBackupConfig"`
	Secret         string `yaml:"Secret" json:"Secret,omitempty"`
}

func (c *config) initApolloConfig() {
	c.ApolloConfig = &apolloConfig{}
	c.ApolloConfig.AppID = c.viper.GetString("apollo.AppID")
	c.ApolloConfig.Cluster = c.viper.GetString("apollo.Cluster")
	c.ApolloConfig.NamespaceName = c.viper.GetString("apollo.NamespaceName")
	c.ApolloConfig.Address = c.viper.GetString("apollo.Address")
	c.ApolloConfig.IsBackupConfig = c.viper.GetBool("apollo.IsBackupConfig")
	c.ApolloConfig.Secret = c.viper.GetString("apollo.Secret")
}

func initApolloClient() {
	var err error
	apolloClient, err = agollo.StartWithConfig(func() (*agollo_config.AppConfig, error) {
		return &agollo_config.AppConfig{
			AppID:          C.ApolloConfig.AppID,
			Cluster:        C.ApolloConfig.Cluster,
			NamespaceName:  C.ApolloConfig.NamespaceName,
			IP:             C.ApolloConfig.Address,
			IsBackupConfig: C.ApolloConfig.IsBackupConfig,
			Secret:         C.ApolloConfig.Secret,
		}, nil
	})
	if err == nil {
		zap.S().Infof("init apollo client success,apollo config: %#v", C.ApolloConfig)
	}
	return
}

func (c *config) GetConfigFromApollo() error {
	cache := apolloClient.GetConfigCache(c.ApolloConfig.NamespaceName)
	value, err := cache.Get("permission_manager_config")
	if err != nil {
		return err
	}
	//zap.S().Info("permission_manager_config:", value)
	C.TimerMiddlewareConfig = &timerMiddlewareConfig{}
	C.TimerMiddlewareConfig.PullDatabaseConfig = make([]*databaseConfig, 0)
	C.TimerMiddlewareConfig.PushDatabaseConfig = make([]*databaseConfig, 0)
	conf := make(map[string]interface{})
	if err = json.Unmarshal([]byte(value.(string)), &conf); err != nil {
		zap.S().Panic("json unmarshal panic:", err)
	}
	zap.S().Infof("config %#v", conf)
	for _, databasesInfo := range conf[constants.PullDatabaseKey].([]interface{}) {
		var DatabaseConfig databaseConfig
		databaseInfoMap := databasesInfo.(map[string]interface{})
		jsonDatabaseInfo, err := json.Marshal(databaseInfoMap)
		if err != nil {
			zap.S().Panic(err)
		}
		if err := json.Unmarshal(jsonDatabaseInfo, &DatabaseConfig); err != nil {
			zap.S().Panic(err)
		}
		C.TimerMiddlewareConfig.PullDatabaseConfig = append(C.TimerMiddlewareConfig.PullDatabaseConfig, &DatabaseConfig)
	}
	for _, databasesInfo := range conf[constants.PushDatabaseKey].([]interface{}) {
		var DatabaseConfig databaseConfig
		databaseInfoMap := databasesInfo.(map[string]interface{})
		jsonDatabaseInfo, err := json.Marshal(databaseInfoMap)
		if err != nil {
			zap.S().Panic(err)
		}
		if err := json.Unmarshal(jsonDatabaseInfo, &DatabaseConfig); err != nil {
			zap.S().Panic(err)
		}
		C.TimerMiddlewareConfig.PushDatabaseConfig = append(C.TimerMiddlewareConfig.PushDatabaseConfig, &DatabaseConfig)
	}
	timeInterval := conf[constants.TimeIntervalKey].(string)
	timeIntervalDuration, err := time.ParseDuration(timeInterval)
	if err != nil {
		zap.S().Panic(err)
	}
	C.TimerMiddlewareConfig.TimeInterval = timeIntervalDuration
	zap.S().Infof("get apollo config from apollo,config:%#v", C.TimerMiddlewareConfig)
	return nil
}
