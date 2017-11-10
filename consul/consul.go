package consul

import (
	"github.com/hashicorp/consul/api"
	"fmt"
	"os"
	"log"
)

var consulClient, _ = api.NewClient(&api.Config{
	// TODO use config
	Address: "192.168.33.34:8500",
})

func Register(consulService ConsulService) error {

	fabioTag := fmt.Sprintf("urlprefix-/%s strip=/%s", consulService.ServiceName, consulService.ServiceName)
	if len(consulService.ServiceSource) > 0 {
		fabioTag = fmt.Sprintf("urlprefix-/%s", consulService.ServiceSource)
	}
	tags := []string{fabioTag, "dev"}
	consulClient.Agent().ServiceDeregister(consulService.ServiceID)
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
	return err
}

func AgentService() (map[string]*api.AgentService, error) {
	return consulClient.Agent().Services()
}
