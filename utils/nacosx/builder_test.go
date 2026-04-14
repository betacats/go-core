package nacosx

import (
	"fmt"
	"testing"
	"time"
)

//type ZeroConfig struct {
//	rest.RestConf
//}
//
//func TestNacos(t *testing.T) {
//	nacosx := NewBuilder().
//		WithServerAddr("localhost", 8848).
//		WithNamespace("public").
//		WithAuth("USER", "PASSWORD").
//		WithDataID("vet").
//		Execute()
//	fmt.Println(nacosx.config.IPAddr)
//
//	var c ZeroConfig
//	nacosx.MustLoad(&c)
//
//	fmt.Println("xxxxx", c.Name)
//
//	nacosx.ListenConfig(func(namespace, group, dataID, data string) {
//		nacosx.MustLoad(&c)
//		fmt.Println("lister....c.Name...", c.Name)
//	})
//
//	time.Sleep(30 * time.Second)
//}

//localhost
//read config from both server and cache fail, err=read cache file Config Encrypted Data Key failed. cause file doesn't exist, file path: cache/nacos/config/vet@@DEFAULT_GROUP@@public.: file not exist，dataId=vet, group=DEFAULT_GROUP, namespaceId=public
//xxxxx

//localhost
//read config from both server and cache fail, err=read cache file Config Encrypted Data Key failed. cause file doesn't exist, file path: cache/nacos/config/vet@@DEFAULT_GROUP@@public.: file not exist，dataId=vet, group=DEFAULT_GROUP, namespaceId=public
//xxxxx

// AppConfig 应用配置示例
type AppConfig struct {
	Name    string `yaml:"Name"`
	Host    string `yaml:"Host"`
	Port    int    `yaml:"Port"`
	Mode    string `yaml:"Mode"`
	Timeout int    `yaml:"Timeout"`
}

func TestNacosByAppConfig(t *testing.T) {
	nacosx := NewBuilder().
		WithServerAddr("localhost", 8848).
		WithNamespace("public").
		WithAuth("USER", "PASSWORD").
		WithDataID("vet").
		Execute()
	fmt.Println(nacosx.config.IPAddr)

	var c AppConfig
	nacosx.MustLoad(&c)

	fmt.Println("xxxxx", c.Name)

	nacosx.ListenConfig(func(namespace, group, dataID, data string) {
		nacosx.MustLoad(&c)
		fmt.Println("lister....c.Name...", c.Name)
	})

	time.Sleep(30 * time.Second)
}
