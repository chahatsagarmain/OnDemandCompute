package main

import (
	"fmt"
	"log"

	"github.com/chahatsagarmain/OnDemandCompute/internal/allocator"
	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
)

func main() {
	resource := manager.AvailableResource{}
	// The timer for resource computation will start with call of Print Resource
	manager.PrintResource(resource)
	client , err := runner.InitDockerClient()
	if err != nil {
		log.Fatal("fatal error : client not started")
	}
	defer client.Client.Close()
	err = client.PullSSHEnabledUbunutImage()
	if err != nil {
		log.Fatal("fatal error : image pull failed")
	}
	allocator , err := allocator.NewAllocator(client)
	if err != nil {
		log.Fatalf("error assigning a new resource allocator %v",allocator)
	}
	_ , err = allocator.AllocateResource("2223",rtypes.Unit{
		MemRequired: 2 * 1024 * 1024 * 1024,
		DiskRequired: 10 * 1024 * 1024 * 1024,
		CpuRequired: 2,
	})
	if err != nil {
		log.Fatal(err)
		log.Fatal("fatal error : couldnt allocate resource")
	}
	containerList , err := client.GetContainerList()
	if err != nil {
		log.Fatalf("error getting container list")
	}
	for _ , val := range(containerList) {
		fmt.Printf("%v",val)
	}
}