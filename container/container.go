package container

import (
	"github.com/zhangkesheng/registrator/consul"
	"github.com/docker/docker/client"
	"context"
	"log"
	"strings"
	"github.com/docker/docker/api/types"
	"strconv"
	"fmt"
	"github.com/zhangkesheng/registrator/weave"
	"errors"
)

var cli, _ = client.NewEnvClient()
var ctx = context.Background()

func GetConsulService(containerID string) (service consul.ConsulService, err error) {
	containerDetails, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Printf("get container details error.%s", err.Error())
		return consul.ConsulService{}, err
	}
	envMap := GetContainerEnvMap(containerDetails)
	servicePort := DEFAULT_PORT
	serviceIp := ""
	if envMap[SERVICE_PORT] != "" {
		servicePort, err = strconv.Atoi(envMap[SERVICE_PORT])
		if err != nil {
			log.Println("Illegal service port")
			servicePort = DEFAULT_PORT
		}
	}
	if envMap[WEAVE_CIDR] != "" {
		// check WEAVE_CIDR attach status
		if !weave.CheckWeaveCIDRAttached(envMap[WEAVE_CIDR]) {
			return consul.ConsulService{}, errors.New(fmt.Sprintf("WEAVE_CIDR [%s] is attached.", envMap[WEAVE_CIDR]))
		}
		serviceIp = envMap[WEAVE_CIDR]
	}
	healthUrl := fmt.Sprintf("%s/%s", servicePort, envMap[HEALTH_CHECK_URL])
	return consul.ConsulService{
		ServiceIp:     serviceIp,
		ServiceID:     containerDetails.ID,
		ServiceName:   envMap[SERVICE_NAME],
		ServicePort:   servicePort,
		HealthUrl:     healthUrl,
		ServiceSource: envMap[SERVICE_SOURCE],
	}, nil
}

func GetContainerEnvMap(container types.ContainerJSON) map[string]string {
	envMap := make(map[string]string)
	config := container.Config
	configEnv := config.Env
	log.Printf("configEnv: %s", configEnv)
	if configEnv != nil && len(configEnv) > 0 {
		for _, v := range configEnv {
			index := strings.Index(v, "=")
			log.Printf("config %s: ", v)
			if index < 0 {
				envMap[v] = ""
			} else if (index + 1) == len(v) {
				envMap[v[:index]] = ""
			} else {
				envMap[v[:index]] = v[index+1:]
			}
		}
	}
	return envMap
}
