package allocate_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/chahatsagarmain/OnDemandCompute/internal/allocator"
	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
	"github.com/chahatsagarmain/OnDemandCompute/proto-gen/github.com/chahatsagarmain/OnDemandCompute/proto-gen/message"
)

const (
	minMemrequired  = 1 * 1024 * 1024 * 1024
	minDiskrequired = 50 * 1024 * 1024 * 1024
	minCpurequired  = 1
)

type AllocateService struct {
	allocator *allocator.Allocator
	message.UnimplementedResourceServiceServer
}

func NewAllocateService(allocator *allocator.Allocator) *AllocateService {
	return &AllocateService{
		allocator: allocator,
	}
}

func (a *AllocateService) AllocateResource(_ context.Context, res *message.ResourceReq) (*message.ResourceRes, error) {
	resourceRreq := rtypes.Unit{
		MemRequired:  res.MemRequired,
		DiskRequired: res.DiskRequired,
		CpuRequired:  int(res.CpuRequired),
	}
	if resourceRreq.MemRequired == 0 {
		resourceRreq.MemRequired = minMemrequired
	}
	if resourceRreq.DiskRequired == 0 {
		resourceRreq.DiskRequired = minDiskrequired
	}
	if resourceRreq.CpuRequired == 0 {
		resourceRreq.CpuRequired = minCpurequired
	}
	var portMappings []runner.PortMapping
    for _, pm := range res.TargetPort {
        portMappings = append(portMappings, runner.PortMapping{
            HostPort:      pm.HostPort,
            ContainerPort: pm.ContainerPort,
        })
    }
	var sshPort string
	for _ , val := range(portMappings){
		if(val.ContainerPort == "22"){
			sshPort = val.HostPort
			break
		}
	}
	if sshPort == "" {
		return &message.ResourceRes{
			Done:    false,
			Message: errors.New("no ssh port specified").Error(),
		}, errors.New("no ssh port specified")
	}
	fmt.Printf("%v", resourceRreq)
	allocated, err := a.allocator.AllocateResource(portMappings, resourceRreq)
	if err != nil {
		return &message.ResourceRes{
			Done:    false,
			Message: err.Error(),
		}, err
	}

	if !allocated {
		return &message.ResourceRes{
			Done:    false,
			Message: "Requested resource couldnt be allocated",
		}, nil
	}

	return &message.ResourceRes{
		Done:    true,
		Message: fmt.Sprintf("Allocated resource : use ssh root@localhost -p %v to connect to the instance", sshPort),
	}, nil
}

func (a *AllocateService) DeleteAllocatedResource(_ context.Context, res *message.ContainerId) (*message.ResourceRes, error) {
	containerId := res.Id
	fmt.Printf("%v", containerId)
	err := a.allocator.DeleteResource(containerId)
	if err != nil {
		return &message.ResourceRes{
			Done:    false,
			Message: err.Error(),
		}, err
	}
	return &message.ResourceRes{
		Done:    true,
		Message: "Deleted Resource allocation",
	}, nil
}

func (a *AllocateService) GetAllocatedResources(_ context.Context, _ *message.Empty) (*message.ContainerInfoRes, error) {
	containerInfo, err := a.allocator.GetResources()
	if err != nil {
		return &message.ContainerInfoRes{
			Containers: make([]*message.ContainerInfo, 0),
		}, err
	}

	return &message.ContainerInfoRes{
		Containers: convertContainerInfo(containerInfo),
	}, nil
}

func (a *AllocateService) GetContainerStats(_ context.Context, cId *message.ContainerId) (*message.ContainerStatsRes, error) {
	containerId := cId.Id
	resp, err := a.allocator.GetResourceStat(containerId)
	if err != nil {
		return &message.ContainerStatsRes{
			ContainerStats: "",
		}, err
	}
	return &message.ContainerStatsRes{
		ContainerStats: resp,
	}, nil
}

func convertContainerInfo(c []runner.ContainerInfo) []*message.ContainerInfo {
	res := make([]*message.ContainerInfo, len(c))
	for idx, val := range c {
		res[idx] = &message.ContainerInfo{
			ContainerId: val.ContainerId,
			State:       val.State,
			Status:      val.Status,
			Image:       val.Image,
			ImageId:     val.ImageId,
			Ports:       convertPorts(val.Ports),
		}
	}
	return res
}

func convertPorts(p []runner.Port) []string {
	ports := make([]string, len(p))
	for idx, val := range p {
		ports[idx] = val.ToString()
	}
	return ports
}
