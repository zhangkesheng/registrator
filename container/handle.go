package container

import (
	"log"
	"github.com/zhangkesheng/registrator/weave"
	"github.com/docker/docker/api/types/events"
	"github.com/zhangkesheng/registrator/consul"
	"strings"
	"github.com/zhangkesheng/registrator/config"
)

func EventsHandler(message events.Message) {
	log.Printf("event message: %s, %s, %s, %s;\n", message.ID, message.Action, message.From, message.Type)
	if message.ID == "" || message.Type != "container" || checkContainerIgnore(message.From) {
		return
	}
	log.Printf("event message: %s, %s, %s, %s;\n", message.ID, message.Action, message.From, message.Type)
	if "kill" == message.Action {
		containerIp := weave.Detach(message.ID)
		log.Printf("weave ip [%s] detach.\n", containerIp)
		_, err := GetConsulService(message.ID)
		if err != nil {
			return
		}
		consul.DeRegister(message.ID)
		return
	}
	if "start" == message.Action {
		consulService, err := GetConsulService(message.ID)
		if err != nil {
			log.Println(err)
			return
		}
		consulService.ServiceIp = weave.Attach(consulService)
		consul.Register(consulService)
	}
}

func EventsErrorHandler(error error) {
	panic(error.Error())
}

func checkContainerIgnore(from string) bool {
	envIgnoreContainers := config.CommonCfg.IgnoreContainer
	if len(envIgnoreContainers) == 0 {
		return false
	}
	if strings.Contains(envIgnoreContainers, ",") {
		containers := strings.Split(envIgnoreContainers, ",")
		for _, v := range containers {
			if strings.Contains(from, v) {
				return true
			}
		}
	} else if strings.Contains(from, envIgnoreContainers) {
		return true
	}
	return false
}
