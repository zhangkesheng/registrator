package weave

import (
	"os/exec"
	"github.com/zhangkesheng/registrator/services"
	"fmt"
	"strings"
	"github.com/zhangkesheng/registrator/consul"
)

func GetWeaveMap() (map[string]string, error) {
	out, err := exec.Command("weave", "ps").CombinedOutput()
	cmdResult := services.CmdResultHandler(out, err)
	result := make(map[string]string)
	for _, v := range cmdResult {
		fmt.Println("=====", v)
		cmdArr := strings.Split(v, " ")
		if len(cmdArr) == 3 {
			result[cmdArr[0]] = cmdArr[2]
		}
	}
	return result, nil
}

func WeaveAttach(service consul.ConsulService) string {
	return ""
}
