package consul

import (
	"github.com/hashicorp/consul/api"
	"fmt"
	"log"
	"errors"
	"github.com/zhangkesheng/registrator/config"
)

var client *api.Client

func InitConsul() {
	client, _ = api.NewClient(&api.Config{
		Address: config.CommonCfg.ConsulHttpAddr,
	})
}

func Register(consulService ConsulService) error {
	if len(consulService.ServiceName) == 0 {
		return errors.New("consul register error. missing params ServiceName")
	}
	if len(consulService.ServiceIp) == 0 {
		return errors.New("consul register error. missing params ServiceIp")
	}
	if len(consulService.HealthUrl) == 0 {
		return errors.New("consul register error. missing params HealthUrl")
	}
	fabioTag := fmt.Sprintf("urlprefix-/%s strip=/%s", consulService.ServiceName, consulService.ServiceName)
	if len(consulService.ServiceSource) > 0 {
		fabioTag = fmt.Sprintf("urlprefix-/%s", consulService.ServiceSource)
	}
	tags := []string{fabioTag, "dev"}
	DeRegister(consulService.ServiceID)
	consulError := client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      consulService.ServiceID,
		Name:    consulService.ServiceName,
		Address: consulService.ServiceIp,
		Port:    consulService.ServicePort,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%s", consulService.ServiceIp, consulService.HealthUrl),
			Interval: config.CommonCfg.ConsulInternal,
		},
	})
	if consulError != nil {
		log.Printf("register error: %s", consulError.Error())
		return consulError
	}
	log.Printf("container [%s] register consul success.", consulService.ServiceName)
	return nil
}

func DeRegister(serviceID string) error {
	err := client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		log.Printf("consul deregister error: %s", err.Error())
	}
	return err
}

func AgentService() (map[string]*api.AgentService, error) {
	return client.Agent().Services()
}
