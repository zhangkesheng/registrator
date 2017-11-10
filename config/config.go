package config

import "os"

type CommonConf struct {
	ConsulInternal  string
	WeaveHostCidr   string
	ConsulHttpAddr  string
	IgnoreContainer string
}

var CommonCfg *CommonConf

func getDefaultConfig() *CommonConf {
	return &CommonConf{
		ConsulInternal:  "10s",
		WeaveHostCidr:   "192.168.195.253/22",
		ConsulHttpAddr:  "192.168.33.34:8500",
		IgnoreContainer: "docker,weave",
	}
}

func LoadCommonConf() *CommonConf {
	conf := getDefaultConfig()
	if os.Getenv("CONSUL_INTERNAL") != "" {
		conf.ConsulInternal = os.Getenv("CONSUL_INTERNAL")
	}
	if os.Getenv("WEAVE_HOST_CIDR") != "" {
		conf.WeaveHostCidr = os.Getenv("WEAVE_HOST_CIDR")
	}
	if os.Getenv("CONSUL_HTTP_ADDR") != "" {
		conf.ConsulHttpAddr = os.Getenv("CONSUL_HTTP_ADDR")
	}
	if os.Getenv("IGNORE_CONTAINER") != "" {
		conf.IgnoreContainer = os.Getenv("IGNORE_CONTAINER")
	}
	return conf
}
