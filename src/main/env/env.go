package env

const (
	ignoreContainer = "daocloud,docker,weave"
)

var envMap map[string]string

func Init() {
	envMap = make(map[string]string)
	envMap["IGNORE_CONTAINER"] = ignoreContainer
	envMap["SERVICE_CHECK_TIMES_MAX"] = "100"
	envMap["CONSUL_HTTP_ADDR"] = ""
	envMap["WEAVE_ROUTER_CMD"] = "--ipalloc-range 192.168.192.0/22 192.168.10.38"
	envMap["WEAVE_HOST_CIDR"] = "192.168.195.253/22"
	envMap["DEFAULT_CONSUL_INTERVAL"] = "10s"
}
