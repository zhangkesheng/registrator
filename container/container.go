package container

import (
	"github.com/zhangkesheng/registrator/consul"
	"github.com/docker/docker/client"
	"context"
	"log"
	"strings"
	"github.com/docker/docker/api/types"
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
	if envmap[SERVICE_PORT] != "" {
		//todo
	}

	//todo 参数校验和初始化
	return consul.ConsulService{
		ServiceID:     containerDetails.ID,
		ServiceName:   envmap[SERVICE_NAME],
		ServicePort:   servicePort,
		HealthUrl:     envmap[HEALTH_CHECK_URL],
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
