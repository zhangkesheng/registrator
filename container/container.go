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
)

var cli, _ = client.NewEnvClient()
var ctx = context.Background()

func GetConsulService(containerID string) (service consul.ConsulService, err error) {
	containerDetails, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		println(err)
	}
	envmap := GetContainerEnvMap(containerDetails)
	servicePort := DEFAULT_PORT;
	serviceIp := ""
	if envmap[SERVICE_PORT] != "" {
		servicePort, err = strconv.Atoi(envmap[SERVICE_PORT])
		if err != nil {
			log.Println("Illegal service port")
			servicePort = DEFAULT_PORT
		}
	}
	if envmap[WEAVE_CIDR] != "" {
		serviceIp = strings.Split(envmap[WEAVE_CIDR], "/")[0]
	}
	healthUrl := fmt.Sprintf("%s/%s", servicePort, envmap[HEALTH_CHECK_URL])
	return consul.ConsulService{
		ServiceIp:     serviceIp,
		ServiceID:     containerDetails.ID,
		ServiceName:   envmap[SERVICE_NAME],
		ServicePort:   servicePort,
		HealthUrl:     healthUrl,
		ServiceSource: envmap[SERVICE_SOURCE],
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
			envMap[v[:index]] = v[index+1:]
		}
	}
	return envMap
}
