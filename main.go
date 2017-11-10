package main

import (
	"github.com/docker/docker/api/types"
	"github.com/zhangkesheng/registrator/config"
	"context"
	"github.com/docker/docker/client"
	"github.com/zhangkesheng/registrator/weave"
	"github.com/zhangkesheng/registrator/consul"
	"log"
	"github.com/zhangkesheng/registrator/container"
)

var ctx = context.Background()
var cli, dockerClientError = client.NewEnvClient()

func main() {
	config.CommonCfg = config.LoadCommonConf()
	consul.InitConsul()
	if dockerClientError != nil {
		panic(dockerClientError)
	}
	go initDockerContainer(ctx, *cli)
	channel, chanErrs := cli.Events(ctx, types.EventsOptions{
	})
	for {
		select {
		case e := <-chanErrs:
			container.EventsErrorHandler(e)
		case m := <-channel:
			go container.EventsHandler(m)
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
	for _, c := range list {
		if c.State != "running" {
			continue
		}
		containerID := c.ID
		consulService, err := container.GetConsulService(containerID)
		if err != nil {
			log.Println(err)
		}
		if consulMap[containerID] != nil {
			if weaveMap[containerID] != "" {
				continue
			}
		}
		if weaveMap[containerID] == "" {
			weaveMap[containerID] = weave.Attach(consulService)
		}
		consulService.ServiceIp = weaveMap[containerID]
		err = consul.Register(consulService)
		if err != nil {
			log.Printf(err.Error())
		}
	}
}
