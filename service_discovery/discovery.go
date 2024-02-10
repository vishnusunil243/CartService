package servicediscovery

import (
	"fmt"
	"log"

	consulapi "github.com/hashicorp/consul/api"
)

const (
	port      = 8083
	serviceId = "cart-service"
)

func RegisterService() {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	addr := "localhost"
	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceId,
		Name:    "cart-server",
		Port:    port,
		Address: addr,
		Check: &consulapi.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d/%s", addr, port, serviceId),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}
	regiErr := consul.Agent().ServiceRegister(registration)
	if regiErr != nil {
		log.Fatal("failed to register service ", regiErr)
	} else {
		log.Printf("successfully registered service %s:%d", addr, port)
	}
}
