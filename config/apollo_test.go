package config

import (
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	agollo_config "github.com/apolloconfig/agollo/v4/env/config"
	_ "github.com/pkg/errors"
	"testing"
)

func TestApollo(t *testing.T) {
	c := &agollo_config.AppConfig{
		AppID:          "permission-manager",
		Cluster:        "dev01",
		NamespaceName:  "tec-do2.0.permission_manager",
		IP:             "http://172.24.2.121:8080",
		IsBackupConfig: true,
		Secret:         "39bfbb6f1d69424ab02ba6f5da265386",
	}

	client, err := agollo.StartWithConfig(func() (*agollo_config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("初始化Apollo配置成功")

	//Use your apollo key to test
	conf := client.GetConfig(c.NamespaceName)
	value := conf.GetStringValue("permission_manager_config", "")
	t.Log(value)
}

func TestInitApolloClient(t *testing.T) {
	if apolloClient == nil {
		t.Error("apolloClient==nil")
	}
	conf := apolloClient.GetConfig(C.ApolloConfig.NamespaceName)
	value := conf.GetStringValue("Test", "")
	t.Log(value)
}

func TestUnmarshalApolloConfig(t *testing.T) {

}
