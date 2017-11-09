package main

import (
	"github.com/docker/docker/api/types"
	"github.com/zhangkesheng/registrator/src/main/config"
	"context"
	"github.com/docker/docker/client"
	"github.com/zhangkesheng/registrator/src/main/services"
)

var ctx = context.Background()
var cli, dockerClientError = client.NewEnvClient()

func main() {
	config.Set_env()
	if dockerClientError != nil {
		panic(dockerClientError)
	}
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
