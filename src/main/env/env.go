package env

const (
	ignoreContainer = "daocloud,docker,weave"
)

var envMap map[string]string

func init() {
	envMap = make(map[string]string)
	envMap["IGNORE_CONTAINER"] = ignoreContainer
}
