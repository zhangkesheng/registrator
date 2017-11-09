package consul

type ConsulService struct {
	ServiceID     string
	HealthUrl     string
	ServiceIp     string
	ServiceName   string
	ServiceSource string
	ServicePort   int
}
