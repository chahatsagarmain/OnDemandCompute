package allocator

import (
	"github.com/chahatsagarmain/OnDemandCompute/internal/runner"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
)

type Allocator struct {
	dockerClient               *runner.DockerClient
	memoryManager              *manager.AvailableResource
	TotalMemory                uint64
	TotalDiskSize              uint64
	TotalCpu                   int
	TotalAvailableCpu          int
	AllocatedContainerMemory   uint64
	AllocatedContainerDiskSize uint64
	ActiveContainers           uint8
	MaxContainers              uint8
	RunningContainer           map[string]ContainerInfo
}

type ContainerInfo struct {
	CotainerId   string
	MemReserved  uint64
	DiskReserved uint64
	CpuReserved  uint64
}

func NewAllocator(dc *runner.DockerClient) (*Allocator, error) {
	mManager := &manager.AvailableResource{}
	totalRam, err := mManager.GetTotalMemSize()
	if err != nil {
		return nil, err
	}
	totalDisk, err := mManager.GetTotalDiskSize()
	if err != nil {
		return nil, err
	}
	totalCpu, err := mManager.GetCpuCount()
	if err != nil {
		return nil, err
	}
	return &Allocator{
		memoryManager:              mManager,
		dockerClient:               dc,
		TotalMemory:                totalRam,
		TotalDiskSize:              totalDisk,
		TotalCpu:                   totalCpu,
		TotalAvailableCpu:          totalCpu,
		ActiveContainers:           0,
		MaxContainers:              10,
		AllocatedContainerMemory:   0,
		AllocatedContainerDiskSize: 0,
		RunningContainer:           make(map[string]ContainerInfo),
	}, nil
}

func (m *Allocator) AllocateResource(targetPort []runner.PortMapping, resource rtypes.Unit) (bool, error) {
	if m.ActiveContainers+1 >= m.MaxContainers {
		return false, nil
	}
	if m.AllocatedContainerMemory+resource.MemRequired >= m.TotalMemory {
		return false, nil
	}
	if m.AllocatedContainerDiskSize+resource.DiskRequired >= m.TotalDiskSize {
		return false, nil
	}
	if m.TotalAvailableCpu-resource.CpuRequired < 0 {
		return false, nil
	}
	err := m.allocateResourceToContainer(targetPort, resource)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *Allocator) allocateResourceToContainer(targetPort []runner.PortMapping, resource rtypes.Unit) error {
	containerId, err := m.dockerClient.StartSSHContainer(targetPort, resource)
	if err != nil {
		return err
	}
	m.TotalAvailableCpu -= resource.CpuRequired
	m.ActiveContainers += 1
	m.AllocatedContainerDiskSize += resource.DiskRequired
	m.AllocatedContainerMemory += resource.MemRequired
	m.RunningContainer[containerId] = ContainerInfo{
		CotainerId:   containerId,
		MemReserved:  resource.MemRequired,
		DiskReserved: resource.DiskRequired,
		CpuReserved:  uint64(resource.CpuRequired),
	}
	return nil
}

func (m *Allocator) DeleteResource(containerId string) error {
	err := m.dockerClient.StopDockerContainer(containerId)
	if err != nil {
		return err
	}
	err = m.dockerClient.DeleteDockerContainer(containerId)
	if err != nil {
		return err
	}
	containterInfo := m.RunningContainer[containerId]
	delete(m.RunningContainer, containerId)
	m.TotalAvailableCpu += int(containterInfo.CpuReserved)
	m.AllocatedContainerMemory -= containterInfo.MemReserved
	m.AllocatedContainerDiskSize -= containterInfo.DiskReserved
	m.ActiveContainers -= 1
	return nil
}

func (m *Allocator) GetResources() ([]runner.ContainerInfo, error) {
	containerList, err := m.dockerClient.GetContainerList()
	if err != nil {
		return nil, err
	}
	return containerList, err
}

func (m *Allocator) GetResourceStat(containerId string) (string, error) {
	resp, err := m.dockerClient.GetContainerStatus(containerId)
	if err != nil {
		return "", nil
	}
	return resp, nil
}
