package consul

type consulNode struct {
	ContainerID   string
	HealthUrl     string
	ContainerIp   string
	ServiceName   string
	ServiceSource string
	ServicePort   string
}
