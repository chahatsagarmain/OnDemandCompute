package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/chahatsagarmain/OnDemandCompute/internal/allocator"
	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	allocate_service "github.com/chahatsagarmain/OnDemandCompute/internal/services/grpc"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/proto-gen/github.com/chahatsagarmain/OnDemandCompute/proto-gen/message"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := "0.0.0.0:50051"
	httpAddr := "0.0.0.0:8080"
	logger := log.Default()
	
	var wg sync.WaitGroup

	wg.Add(1)
	// Start gRPC server
	go func() {
		defer wg.Done()
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Fatalf("Error starting listener on address %v: %v\n", addr, err)
		}

		server := grpc.NewServer()
		resource := manager.AvailableResource{}
		manager.PrintResource(resource)

		client, err := runner.InitDockerClient()
		if err != nil {
			logger.Fatalln("Fatal error: Docker client not started")
		}
		defer client.Client.Close()

		err = client.PullSSHEnabledUbunutImage()
		if err != nil {
			logger.Fatalln("Fatal error: Image pull failed")
		}

		allocatorInstance, err := allocator.NewAllocator(client)
		if err != nil {
			logger.Fatalf("Error assigning a new resource allocator: %v", err)
		}

		allocatorGRPC := allocate_service.NewAllocateService(allocatorInstance)
		message.RegisterResourceServiceServer(server, allocatorGRPC)

		logger.Printf("gRPC server listening at %v\n", addr)
		if err := server.Serve(listener); err != nil {
			logger.Fatalf("gRPC server stopped: %v\n", err)
		}
	}()

	wg.Add(1)
	// Start HTTP gateway
	go func() {
		defer wg.Done()
		mux := runtime.NewServeMux()

		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatalf("Error creating gRPC client connection: %v", err)
		}
		defer conn.Close()

		if err = message.RegisterResourceServiceHandler(context.Background(), mux, conn); err != nil {
			logger.Fatalf("Error registering resource service handler: %v", err)
		}

		logger.Printf("HTTP gateway listening at %v\n", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			logger.Fatalf("HTTP gateway server stopped: %v\n", err)
		}
	}()
	
	wg.Wait()
}