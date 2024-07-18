package config

import (
	"encoding/json"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
	"log"
	"permission/logs"
	"time"
)

// 能通过配置中心等方式传入sql（或其他）配置定时器拉取其他系统的sql语句灵活查询所需权限数据。
// 预留代码接口方便后续不同平台的api接入，定时器暂定一天执行一次，用于定时拉取接入平台的权限体系拉取到程序数据库中
// 自定义sql拉取数据

var C *config = initConfig()

type config struct {
	viper                 *viper.Viper
	ApolloConfig          *apolloConfig          `yaml:"apollo" json:"apollo,omitempty"`
	TimerMiddlewareConfig *timerMiddlewareConfig `yaml:"timer_middleware" json:"timer_middleware,omitempty"`
	XXLJobConfig          *xxlJobConfig          `yaml:"xxl_job" json:"xxl_job,omitempty"`
}

type timerMiddlewareConfig struct {
	PullDatabaseConfig []*databaseConfig `yaml:"pull_database" json:"pull_database,omitempty"`
	PushDatabaseConfig []*databaseConfig `yaml:"push_database" json:"push_database,omitempty"`
	TimeInterval       time.Duration     `yaml:"time_interval" json:"time_interval,omitempty"`
}

type databaseConfig struct {
	DBName string   `yaml:"db_name" json:"db_name,omitempty"`
	DSN    string   `yaml:"dsn" json:"dsn"` //root:123456@tcp(localhost:3306)/Dbname?charset=utf8&parseTime=True&loc=Local
	Sqls   []string `yaml:"sqls" json:"sqls"`
}

type xxlJobConfig struct {
	AppName      string `yaml:"app_name" json:"app_name"`
	AdminAddress string `yaml:"admin_address" json:"admin_address"`
	Token        string `yaml:"token" json:"token"`
	ClientPort   int    `yaml:"client_ports" json:"client_ports"`
	ExecutorIp   string `yaml:"executor_ip" json:"executor_ip"`
}

func initConfig() *config {
	conf := &config{viper: viper.New()}
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("./config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		errors.Wrap(err)
	}
	// init zap log
	conf.initZapLogConfig()

	// init apollo config
	conf.initApolloConfig()

	// init xxl job config
	conf.initXXLJobConfig()

	configInfo, err := json.Marshal(conf)
	if err != nil {
		zap.S().DPanic(err)
	}
	zap.S().Info("permission-manager config \n", string(configInfo))
	return conf
}

//// 热更新
//func (c *config) watchConfig() {
//	go c.viper.WatchConfig()
//	c.viper.OnConfigChange(func(e fsnotify.Event) {
//		if err := c.viper.Unmarshal(&C); err != nil {
//			zap.S().Panic(err)
//		}
//		initConfig()
//	})
//}

func (c *config) initZapLogConfig() {
	//从配置中读取日志配置，初始化日志
	lc := &logs.LogConfig{
		DebugFileName: c.viper.GetString("log.debugFileName"),
		InfoFileName:  c.viper.GetString("log.infoFileName"),
		WarnFileName:  c.viper.GetString("log.warnFileName"),
		MaxSize:       c.viper.GetInt("log.maxSize"),
		MaxAge:        c.viper.GetInt("log.maxAge"),
		MaxBackups:    c.viper.GetInt("log.maxBackups"),
	}

	err := logs.InitLogger(lc)
	if err != nil {
		log.Fatalln(err)
	}
	zap.S().Infof("init zap log success , log config: %#v", lc)
}

func (c *config) initXXLJobConfig() {
	xc := &xxlJobConfig{
		AppName:      c.viper.GetString("xxl.app_name"),
		AdminAddress: c.viper.GetString("xxl.admin_address"),
		Token:        c.viper.GetString("xxl.token"),
		ClientPort:   c.viper.GetInt("xxl.client_ports"),
		ExecutorIp:   c.viper.GetString("xxl.executor_ip"),
	}
	c.XXLJobConfig = xc
}
