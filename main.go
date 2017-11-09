package main

import (
	"github.com/docker/docker/api/types"
	"github.com/zhangkesheng/registrator/config"
	"context"
	"github.com/docker/docker/client"
	"github.com/zhangkesheng/registrator/weave"
	"github.com/zhangkesheng/registrator/consul"
	"github.com/zhangkesheng/registrator/continer"
	"fmt"
	"log"
	"github.com/zhangkesheng/registrator/services"
)

var ctx = context.Background()
var cli, dockerClientError = client.NewEnvClient()

func main() {
	config.Set_env()
	if dockerClientError != nil {
		panic(dockerClientError)
	}
	go initDockerContainer(ctx, *cli)
	channel, chanErrs := cli.Events(ctx, types.EventsOptions{
	})
	for {
		select {
		case e := <-chanErrs:
			services.EventsErrorHandler(e)
		case m := <-channel:
			go services.EventsHandler(m)
		}
	}
}

func initDockerContainer(ctx context.Context, cli client.Client) {
	log.Printf("init docker container.\n")
	list, err := cli.ContainerList(ctx, types.ContainerListOptions{
	})
	if err != nil {
		println(err)
	}
	consulMap, err := consul.AgentService()
	if err != nil {
		println(err)
	}
	weaveMap, err := weave.GetWeaveMap()
	if err != nil {
		println(err)
	}
	for _, container := range list {
		containerID := container.ID
		consulService, err := continer.GetConsulService(containerID)
		if err != nil {
			fmt.Println(err)
		}

		if container.State != "running" {
			continue
		}
		//当consul中存在此镜像
		if consulMap[containerID] != nil {
			if weaveMap[containerID] != "" {
				//todo 判断ip是否相同
				continue
			}
		}
		if weaveMap[containerID] == "" {
			weaveMap[containerID] = "todo"
		}
		consulService.ServiceIp = weaveMap[containerID]
		consul.Register(consulService)
	}
}
