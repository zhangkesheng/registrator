package consul

import (
	"github.com/hashicorp/consul/api"
	"fmt"
	"os"
	"log"
	"errors"
)

var consulClient, _ = api.NewClient(&api.Config{
	// TODO use config
	Address: "192.168.33.34:8500",
})

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
	consulError := consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      consulService.ServiceID,
		Name:    consulService.ServiceName,
		Address: consulService.ServiceID,
		Port:    consulService.ServicePort,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%s", consulService.ServiceIp, consulService.HealthUrl),
			Interval: os.Getenv("DEFAULT_CONSUL_INTERVAL"),
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
	err := consulClient.Agent().ServiceDeregister(serviceID)
	if err != nil {
		log.Printf("consul deregister error: %s", err.Error())
	}
	return err
}

func AgentService() (map[string]*api.AgentService, error) {
	return consulClient.Agent().Services()
}
