package services

import (
	"fmt"
	"strconv"
	"os"
	"log"
	"github.com/hashicorp/consul/api"
	"github.com/docker/docker/api/types"
)

var consulClient, _ = api.NewClient(&api.Config{
	// TODO use config
	Address: "192.168.33.34:8500",
})

// create consul service
func ConsulCreate(container types.ContainerJSON, envMap map[string]string, containerIp string, healthUrl string) {
	fabioTag := fmt.Sprintf("urlprefix-/%s strip=/%s", envMap[SERVICE_NAME], envMap[SERVICE_NAME])
	if len(envMap[SERVICE_SOURCE]) > 0 {
		fabioTag = fmt.Sprintf("urlprefix-/%s", envMap[SERVICE_SOURCE])
	}
	tags := []string{fabioTag, "dev"}
	consulClient.Agent().ServiceDeregister(container.ID)
	portInt, _ := strconv.Atoi(envMap[SERVICE_PORT])
	consulError := consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      container.ID,
		Name:    container.Name,
		Address: containerIp,
		Port:    portInt,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     healthUrl,
			Interval: os.Getenv("DEFAULT_CONSUL_INTERVAL"),
		},
	})
	if consulError != nil {
		log.Printf("register error: %s", consulError.Error())
	} else {
		log.Printf("container [%s] register consul success.", container.Name)
	}
}

// deregister consul service
func ConsulDeregister(seriviceId string) {
	consulClient.Agent().ServiceDeregister(seriviceId)
}
