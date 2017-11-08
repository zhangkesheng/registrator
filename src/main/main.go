package main

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"context"
	"github.com/docker/docker/api/types/events"
	"fmt"
	"os"
	"strings"
	"os/exec"
	"log"
	"github.com/hashicorp/consul/api"
	"strconv"
	"net/http"
	"github.com/zhangkesheng/registrator/src/main/env"
)

const (
	SERVICE_PORT     = "SERVER_PORT"
	SERVICE_NAME     = "SERVICE_NAME"
	HEALTH_CHECK_URL = "HEALTH_CHECK_URL"
	SERVICE_SOURCE   = "SERVICE_SOURCE"
)

var ctx = context.Background()
var consulClient, consulClientError = api.NewClient(api.DefaultConfig())
var cli, dockerClientError = client.NewEnvClient()

func main() {
	env.Set_env()
	if consulClientError != nil {
		panic(consulClientError)
	}
	if dockerClientError != nil {
		panic(dockerClientError)
	}
	channel, chanErrs := cli.Events(ctx, types.EventsOptions{
	})
	for {
		select {
		case e := <-chanErrs:
			eventsErrorHandler(e)
		case m := <-channel:
			go eventsHandler(m)
		}
	}
}

// handle docker events
func eventsHandler(message events.Message) {
	// message id, message type is not null and type is container, should handle this event
	if len(message.ID) > 0 && len(message.Type) > 0 && message.Type == "container" && !CheckContainerIgnore(message.From) {
		fmt.Printf("event message: %s, %s, %s, %s;\n", message.ID, message.Action, message.From, message.Type)
		// delete weave attach and deregister consul service where container stop
		if "stop" == message.Action {
			out, err := exec.Command("weave", "detach", message.ID).CombinedOutput()
			cmdResult := CmdResultHandler(out, err)
			containerIp := cmdResult[0]
			log.Printf("weave ip [%s] detach.\n", containerIp)
			container := GetContainerConfig(message.ID)
			if len(container.ID) > 0 {
				ConsulDeregister(container.ID)
			} else {
				log.Printf("container [%s] not found.", message.ID)
			}
		} else if "start" == message.Action {
			// weave attach and register consul service where container start
			out, err := exec.Command("weave", "attach", message.ID).CombinedOutput()
			cmdResult := CmdResultHandler(out, err)
			containerIp := cmdResult[0]
			log.Printf("weave ip [%s] attach.\n", containerIp)
			container := GetContainerConfig(message.ID)
			if len(container.ID) > 0 {
				// get container env
				envMap := GetContainerEnvMap(container)
				if CheckContainerEnv(envMap) {
					heathUrl := GetContainerHealthCheckUrl(message.From, containerIp, envMap[HEALTH_CHECK_URL], envMap[SERVICE_PORT])
					log.Printf("service heathUrl: %s", heathUrl)
					if CheckServiceHealthy(heathUrl) {
						fabioTag := fmt.Sprintf("urlprefix-/%s strip=/%s", envMap[SERVICE_NAME], envMap[SERVICE_NAME])
						if len(envMap[SERVICE_SOURCE]) > 0 {
							fabioTag = fmt.Sprintf("urlprefix-/%s", envMap[SERVICE_SOURCE])
						}
						tags := []string{fabioTag, "dev"}
						ConsulCreate(
							container.ID,
							container.Name,
							containerIp,
							envMap[SERVICE_PORT],
							tags,
							heathUrl,
							os.Getenv("DEFAULT_CONSUL_INTERVAL"),
						)
					}
				}
			}
		}
	}
}

// handle docker events error
// TODO send message to admin
func eventsErrorHandler(error error) {
	panic(error.Error())
}

// check container is ignore
func CheckContainerIgnore(from string) bool {
	envIgnoreContainers := os.Getenv("IGNORE_CONTAINER")
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

func CmdResultHandler(out []byte, err error) []string {
	if err != nil {
		log.Printf("weave cmd error: %s", err.Error())
	}
	log.Printf("weave cmd result: %s", string(out))
	resultList := strings.Split(string(out), " \n")
	for i := 0; i < len(resultList); i++ {
		resultList[i] = strings.Replace(resultList[i], "\n", "", -1)
	}
	return resultList
}

// check container healthy
// if response status is 200, it means service is healthy
// if not, check again, the limit times get by env 'SERVICE_CHECK_TIMES_MAX'
func CheckServiceHealthy(healthUrl string) bool {
	serviceCheckTimesMax, _ := strconv.Atoi(os.Getenv("SERVICE_CHECK_TIMES_MAX"))
	for i := 1; i < serviceCheckTimesMax; i++ {
		resp, err := http.Get(healthUrl)
		if err != nil {
			log.Printf("get services healthy [%s] error.%v", healthUrl, err)
		}
		log.Printf("%s times get service healthy. status: %s", i, resp.StatusCode)
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}
	return false
}

// create consul service
func ConsulCreate(containerId string, containerName string, ip string, port string, tags []string, healthUrl string, interval string) {
	consulClient.Agent().ServiceDeregister(containerId)
	portInt, _ := strconv.Atoi(port)
	consulError := consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      containerId,
		Name:    containerName,
		Address: ip,
		Port:    portInt,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     healthUrl,
			Interval: interval,
		},
	})
	log.Printf("register error: %s", consulError.Error())
}

// deregister consul service
func ConsulDeregister(seriviceId string) {
	consulClient.Agent().ServiceDeregister(seriviceId)
}

func GetContainerConfig(containerId string) types.ContainerJSON {
	response, _, err := cli.ContainerInspectWithRaw(ctx, containerId, true)
	if err != nil {
		panic(err)
		return types.ContainerJSON{}
	}
	return response
}

// get container env map by container config
func GetContainerEnvMap(container types.ContainerJSON) map[string]string {
	envMap := make(map[string]string)
	config := container.Config
	configEnv := config.Env
	log.Printf("configEnv: %s", configEnv)
	if configEnv != nil && len(configEnv) > 0 {
		for _, v := range configEnv {
			index := strings.Index(v, "=")
			log.Printf("env %s: ", v)
			envMap[v[:index]] = v[index+1:]
		}
	}
	return envMap
}

// check register env
func CheckContainerEnv(envMap map[string]string) bool {
	log.Printf("service envs: %s\n", envMap)
	if envMap == nil || len(envMap) <= 0 {
		log.Println("services envs is null.")
		return false
	}
	if len(envMap[SERVICE_PORT]) == 0 {
		log.Println("services service port is null.")
		return false
	}
	if _, err := strconv.Atoi(envMap[SERVICE_PORT]); err != nil {
		log.Println("services service port is not int.")
		return false
	}
	if len(envMap[SERVICE_NAME]) == 0 {
		log.Println("services service name is null.")
		return false
	}
	if len(envMap[HEALTH_CHECK_URL]) == 0 {
		log.Println("services health check url is null.")
		return false
	}
	return true
}

func GetContainerHealthCheckUrl(from string, ip string, healthUrl string, port string) string {
	if len(port) == 0 {
		port = "80"
	}
	if strings.Contains(from, "cadvisor") {
		return fmt.Sprintf("http://%s:8080/%s", ip, "healthz")
	} else {
		return fmt.Sprintf("http://%s:%s/%s", ip, port, healthUrl)
	}
}
