package allocate_service

import (
	"context"

	"github.com/chahatsagarmain/OnDemandCompute/internal/allocator"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
	"github.com/chahatsagarmain/OnDemandCompute/proto-gen/github.com/chahatsagarmain/OnDemandCompute/proto-gen/message"
)

type AllocateService struct {
	allocator *allocator.Allocator
	message.UnimplementedResourceServiceServer
}

func NewAllocateService (allocator *allocator.Allocator) (*AllocateService){
	return &AllocateService{
		allocator: allocator,
	}
}

func (a *AllocateService) AllocateResource(_ context.Context ,res *message.ResourceReq) (*message.ResourceRes,error) {
	resourceRreq := rtypes.Unit{
		MemRequired: res.MemRequired,
		DiskRequired: res.DiskRequired,
		CpuRequired: int(res.CpuRequired),
	}
	sshPort := res.SshPort
	allocated , err := a.allocator.AllocateResource(sshPort , resourceRreq)
	if err != nil {
		return &message.ResourceRes{
			Done: false,
			Message: err.Error(),
		} , err
	}

	if !allocated {
		return &message.ResourceRes{
			Done: false,
			Message: "Requested resource couldnt be allocated",
		} , nil
	}

	return &message.ResourceRes{
		Done: true,
		Message: "Allocated resource",
	} , nil
}

func (a *AllocateService) DeleteResource(_ context.Context , res *message.ContainerId) (*message.ResourceRes , error) {
	containerId := res.Id
	err := a.allocator.DeleteResource(containerId)
	if err != nil {
		return &message.ResourceRes{
			Done: false,
			Message: err.Error(),
		} , err
	}
	return &message.ResourceRes{
		Done: true,
		Message: "Deleted Resource allocation",
	} , nil
}