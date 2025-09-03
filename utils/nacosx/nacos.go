package nacosx

import (
	"errors"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	zeroConf "github.com/zeromicro/go-zero/core/conf"
)

type Nacosx struct {
	config       *config
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

// 服务注册相关方法

// RegisterService 注册服务到Nacos
func (s *Nacosx) RegisterService() error {
	success, err := s.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          s.config.ServiceIP,
		Port:        s.config.ServicePort,
		ServiceName: s.config.ServiceName,
		GroupName:   s.config.ServiceGroup,
		ClusterName: s.config.ServiceCluster,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
	})
	if err != nil {
		return err
	}
	if !success {
		return errors.New("service registration failed")
	}
	return nil
}

// DeregisterService 从Nacos注销服务
func (s *Nacosx) DeregisterService() error {
	success, err := s.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          s.config.ServiceIP,
		Port:        s.config.ServicePort,
		ServiceName: s.config.ServiceName,
		GroupName:   s.config.ServiceGroup,
		Cluster:     s.config.ServiceCluster,
	})
	if err != nil {
		return err
	}
	if !success {
		return errors.New("service deregistration failed")
	}
	return nil
}

// 配置管理相关方法

// GetConfig 获取配置
func (s *Nacosx) GetConfig() (string, error) {
	return s.configClient.GetConfig(vo.ConfigParam{
		DataId: s.config.DataID,
		Group:  s.config.ServiceGroup,
	})
}

// PublishConfig 发布配置
func (s *Nacosx) PublishConfig(content string) (bool, error) {
	return s.configClient.PublishConfig(vo.ConfigParam{
		DataId:  s.config.DataID,
		Group:   s.config.ServiceGroup,
		Content: content,
	})
}

// DeleteConfig 删除配置
func (s *Nacosx) DeleteConfig() (bool, error) {
	return s.configClient.DeleteConfig(vo.ConfigParam{
		DataId: s.config.DataID,
		Group:  s.config.ServiceGroup,
	})
}

// ListenConfig 监听配置变化
func (s *Nacosx) ListenConfig(callback func(namespace, group, dataID, data string)) error {
	return s.configClient.ListenConfig(vo.ConfigParam{
		DataId:   s.config.DataID,
		Group:    s.config.ServiceGroup,
		OnChange: callback,
	})
}

// GetServiceInstances 获取服务实例列表
func (s *Nacosx) GetServiceInstances() ([]model.Instance, error) {
	return s.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: s.config.ServiceName,
		GroupName:   s.config.ServiceGroup,
		HealthyOnly: true,
	})
}

func (s *Nacosx) MustLoad(v interface{}) {
	a, err := s.GetConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	zeroConf.LoadConfigFromYamlBytes([]byte(a), v)
}
