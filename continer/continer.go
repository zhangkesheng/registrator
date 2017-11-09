package continer

import "github.com/zhangkesheng/registrator/consul"

func GetConsulService(continerID string) (service consul.ConsulService, err error) {
	return consul.ConsulService{}, nil
}
