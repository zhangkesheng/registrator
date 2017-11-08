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
	"io"
	"bufio"
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
	fmt.Printf("event message: %s, %s, %s, %s;\n", message.ID, message.Action, message.From, message.Type)
	// message id, message type is not null and type is container, should handle this event
	if len(message.ID) > 0 && len(message.Type) > 0 && message.Type == "container" && !CheckContainerIgnore(message.From) {
		// delete weave attach and deregister consul service where container stop
		if "stop" == message.Action {
			cmd := exec.Command("weave", "detach", message.ID)
			cmdResult := CmdResultHandler(cmd)
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
			cmd := exec.Command("weave", "attach", message.ID)
			log.Println(cmd)
			cmdResult := CmdResultHandler(cmd)
			containerIp := cmdResult[0]
			log.Printf("weave ip [%s] detach.\n", containerIp)
			container := GetContainerConfig(message.ID)
			if len(container.ID) > 0 {
				// get container env
				envMap := GetContainerEnvMap(container)
				if CheckContainerEnv(envMap) {
					heathUrl := GetContainerHealthCheckUrl(message.From, containerIp, envMap[HEALTH_CHECK_URL], envMap[SERVICE_PORT])
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
	envIgnoreContainers := os.Getenv("IGNORE_CONTAINERS")
	log.Printf("envIgnoreContainers: %s\n", envIgnoreContainers)
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

func CmdResultHandler(cmd *exec.Cmd) []string {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	cmd.Start()
	printCmdLog(stderr)
	cmd.Wait()
	return GetWeaveCmdResult(stdout)
}

func GetWeaveCmdResult(readCloser io.ReadCloser) []string {
	reader := bufio.NewReader(readCloser)
	var result []string
	var i = 0
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			continue
		}
		log.Printf("weave cmd result: %s\n", line)
		result[i] = line
		i++
		break
	}
	return result
}

func printCmdLog(readCloser io.ReadCloser) {
	reader := bufio.NewReader(readCloser)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Print(line)
	}
}

// check container healthy
// if response status is 200, it means service is healthy
// if not, check again, the limit times get by env 'SERVICE_CHECK_TIMES_MAX'
func CheckServiceHealthy(healthUrl string) bool {
	serviceCheckTimesMax, _ := strconv.Atoi(os.Getenv("SERVICE_CHECK_TIMES_MAX"))
	for i := 0; i < serviceCheckTimesMax; i++ {
		resp, err := http.Get(healthUrl)
		if err != nil {
			log.Printf("get services healthy [%s] error.%v", healthUrl, err)
		}
		if resp.StatusCode == http.StatusOK {
			return true
		}
	}
	return false
}

// create consul service
func ConsulCreate(containerId string, containerName string, ip string, port string, tags []string, healthUrl string, interval string) {
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
	envs := config.Env
	if envs != nil && len(envs) > 0 {
		for _, v := range envs {
			index := strings.Index(v, "=")
			envMap[v[:index]] = v[index+1:]
		}
	}
	return envMap
}

// check register env
func CheckContainerEnv(envMap map[string]string) bool {
	if envMap == nil || len(envMap) <= 0 {
		log.Println("services envs is null.")
		return false
	}
	if len(envMap[SERVICE_PORT]) == 0 {
		log.Println("services service port is null.")
		return false
	}
	if _, err := strconv.Atoi(envMap[SERVICE_PORT]); err == nil {
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
