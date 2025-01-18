package main

import (
	"log"
	"net"

	"github.com/chahatsagarmain/OnDemandCompute/internal/allocator"
	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	allocate_service "github.com/chahatsagarmain/OnDemandCompute/internal/services/grpc"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/proto-gen/github.com/chahatsagarmain/OnDemandCompute/proto-gen/message"
	"google.golang.org/grpc"
)

func main() {
	addr := "0.0.0.0:50051"
	logger := log.Default()
	listener , err := net.Listen("tcp",addr)
	if err != nil {
		logger.Fatalf("Error starting listener on address %v \n" , addr)
	}
	server := grpc.NewServer()
	resource := manager.AvailableResource{}
	// The timer for resource computation will start with call of Print Resource
	manager.PrintResource(resource)
	client , err := runner.InitDockerClient()
	if err != nil {
		logger.Fatalln("fatal error : client not started")
	}
	defer client.Client.Close()
	err = client.PullSSHEnabledUbunutImage()
	if err != nil {
		logger.Fatalln("fatal error : image pull failed")
	}
	allocator , err := allocator.NewAllocator(client)
	if err != nil {
		logger.Fatalf("error assigning a new resource allocator %v",allocator)
	}
	allocator_grpc := allocate_service.NewAllocateService(allocator)
	message.RegisterResourceServiceServer(server , allocator_grpc)	
	logger.Printf("Serving requests at address %v \n",addr)
	if err := server.Serve(listener) ; err != nil {
		logger.Fatalf("Error : Cant serve on addr %v \n", addr)
	}
	
}