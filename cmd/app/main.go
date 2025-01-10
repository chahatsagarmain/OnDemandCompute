package main

import (
	"time"

	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
)

func main() {
	resource := manager.AvailableResource{}
	// The resource computation will start with call of Print Resource
	manager.PrintResource(resource)
	time.Sleep(10 * time.Second)
}