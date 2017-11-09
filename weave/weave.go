package weave

import (
	"os/exec"
	"github.com/zhangkesheng/registrator/services"
	"fmt"
)

func GetWeaveMap() (map[string]string, error) {
	out, err := exec.Command("weave", "ps").CombinedOutput()
	cmdResult := services.CmdResultHandler(out, err)
	fmt.Println(cmdResult)
	result := make(map[string]string)
	return result, nil
}
