package allocator

import (

	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
)

type Allocator struct {
	dockerClient *runner.DockerClient
	memoryManager *manager.AvailableResource
	TotalMemory uint64
	TotalDiskSize uint64
	TotalCpu int
	TotalAvailableCpu int
	AllocatedContainerMemory uint64
	AllocatedContainerDiskSize uint64
	ActiveContainers uint8
	MaxContainers	uint8
}

func NewAllocator(dc *runner.DockerClient) (*Allocator , error) {
	mManager := &manager.AvailableResource{}
	totalRam , err := mManager.GetTotalMemSize()
	if err != nil {
		return nil , err
	}
	totalDisk , err := mManager.GetTotalDiskSize()
	if err != nil {
		return nil , err
	}
	totalCpu , err := mManager.GetCpuCount()
	if err != nil {
		return nil , err
	}
	return &Allocator{
		memoryManager: mManager,
		dockerClient: dc,
		TotalMemory: totalRam,
		TotalDiskSize: totalDisk,
		TotalCpu: totalCpu,
		TotalAvailableCpu: totalCpu,
		ActiveContainers: 0,
		MaxContainers: 10,
		AllocatedContainerMemory: 0,
		AllocatedContainerDiskSize: 0,
	} , nil
}

func(m *Allocator) AllocateResource(sshPort string , resource rtypes.Unit) (bool , error){
	if m.ActiveContainers + 1 >= m.MaxContainers {
		return false , nil
	}
	if m.AllocatedContainerMemory + resource.MemRequired >= m.TotalMemory {
		return false , nil
	}
	if m.AllocatedContainerDiskSize + resource.DiskRequired >= m.TotalDiskSize {
		return false , nil
	}
	if m.TotalAvailableCpu - resource.CpuRequired < 0 {
		return false , nil
	} 
	err := m.allocateResourceToContainer(sshPort , resource)
	if err != nil {
		return false , err
	}
	return true , nil
}

func(m *Allocator) allocateResourceToContainer(sshPort string , resource rtypes.Unit) (error) {
	err := m.dockerClient.StartSSHContainer(sshPort , resource)
	if err != nil{
		return err
	}
	m.TotalAvailableCpu -= resource.CpuRequired
	m.ActiveContainers += 1
	m.AllocatedContainerDiskSize += resource.DiskRequired
	m.AllocatedContainerMemory += resource.MemRequired
	return nil
}