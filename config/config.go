package config

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
	//todo
	return getDefaultConfig()
}
