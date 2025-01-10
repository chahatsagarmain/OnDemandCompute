package main

import (
	"fmt"
	"log"

	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
)

func main() {
	resource := manager.AvailableResource{}
	// The resource computation will start with call of Print Resource
	manager.PrintResource(resource)
	client , err := runner.InitDockerClient()
	if err != nil {
		log.Fatal("fatal error : client not started")
	}
	err = client.PullSSHEnabledUbunutImage()
	if err != nil {
		log.Fatal("fatal error : image pull failed")
	}
	err = client.StartSSHContainer("2222")
	if err != nil {
		log.Fatal("error starting container")
	}
	containerList , err := client.GetContainerList()
	if err != nil {
		log.Fatalf("error getting container list")
	}
	for _ , val := range(containerList) {
		fmt.Printf("%v",val)
	}
}