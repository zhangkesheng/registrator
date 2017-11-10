package weave

import (
	"os/exec"
	"strings"
	"log"
	"github.com/zhangkesheng/registrator/consul"
)

func GetWeaveMap() (map[string]string, error) {
	out, err := exec.Command("weave", "ps").CombinedOutput()
	cmdResult := cmdResultHandler(out, err)
	result := make(map[string]string)
	for _, v := range cmdResult {
		cmdArr := strings.Split(v, " ")
		if len(cmdArr) == 3 {
			result[cmdArr[0]] = cmdArr[2]
		}
	}
	return result, nil
}

func Attach(consulService consul.ConsulService) string {
	if consulService.ServiceIp == "" {
		out, err := exec.Command("weave", "attach", consulService.ServiceID).CombinedOutput()
		cmdResult := cmdResultHandler(out, err)
		return cmdResult[0]
	}
	out, err := exec.Command("weave", "attach", consulService.ServiceID, consulService.ServiceIp).CombinedOutput()
	cmdResult := cmdResultHandler(out, err)
	return cmdResult[0]
}

func Detach(containerID string) string {
	out, err := exec.Command("weave", "detach", containerID).CombinedOutput()
	cmdResult := cmdResultHandler(out, err)
	return cmdResult[0]
}
func cmdResultHandler(out []byte, err error) []string {
	if err != nil {
		log.Printf("weave cmd error: %s", err.Error())
	}
	log.Printf("weave cmd result: %s", string(out))
	resultList := strings.Split(string(out), "\n")
	for i := 0; i < len(resultList); i++ {
		resultList[i] = strings.Replace(resultList[i], "\n", "", -1)
	}
	return resultList
}

func CheckWeaveCIDRAttached(ip string) bool {
	weaveMap, err := GetWeaveAttachIpList()
	if err != nil {
		log.Printf("weave ps error. message: %s", err.Error())
	}
	if len(weaveMap[ip]) > 0 {
		return true
	}
	return false
}

func GetWeaveAttachIpList() (map[string]string, error) {
	out, err := exec.Command("weave", "ps").CombinedOutput()
	cmdResult := cmdResultHandler(out, err)
	result := make(map[string]string)
	for _, v := range cmdResult {
		cmdArr := strings.Split(v, " ")
		if len(cmdArr) == 3 {
			result[cmdArr[2]] = cmdArr[0]
		}
	}
	return result, nil
}
