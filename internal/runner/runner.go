package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/chahatsagarmain/OnDemandCompute/pkg/manager"
	"github.com/chahatsagarmain/OnDemandCompute/pkg/rtypes"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	sshImage = "rastasheep/ubuntu-sshd:latest"
)

type DockerClient struct {
	Client *client.Client
}

type ContainerInfo struct {
	containerId string 
	state 		string
	status 		string
	image 		string
	imageId		string
	ports 		[]Port
}

type Port struct {
	portIP string
	privatePort uint16
	publicPort uint16
	portType string
}

func InitDockerClient() (*DockerClient , error) {
	apiClient , err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil , err
	}
	return &DockerClient{
		Client: apiClient,
	} , nil
}

func (c DockerClient) PullSSHEnabledUbunutImage() (error){
	
	fmt.Println("pulling docker image")
	reader , err := c.Client.ImagePull(context.Background() , sshImage , image.PullOptions{})
	if err != nil {
		fmt.Println("Error pulling image:", err)
		return err
	}
	decoder := json.NewDecoder(reader)
	for {
		var message map[string]interface{}
        if err := decoder.Decode(&message); err != nil {
            if err == io.EOF {
                break
            }
            return err
        }
        if status, ok := message["status"]; ok {
            fmt.Printf("\r%v\n", status)
        }
	}
	return nil
}

func (c DockerClient) StartSSHContainer(sshPort string,requiredResource rtypes.Unit) (error){
	// default password is root
	err := manager.CheckPortAvailable(sshPort)
	if err != nil {
		return fmt.Errorf("PORT %v is already taken",sshPort)
	}
	portBindings := nat.PortMap{
        "22/tcp": []nat.PortBinding{
            {
                HostIP:   "0.0.0.0",
                HostPort: sshPort,
            },
        },
    }
	containerConfig := &container.Config{
        Image: sshImage,
        ExposedPorts: nat.PortSet{
            "22/tcp": struct{}{},
        },
    }	

	// Storage limit is disabled for now because it needs enabling of 'pquota' on local system

    hostConfig := &container.HostConfig{
        PortBindings: portBindings,
		Resources: container.Resources{
			Memory: int64(requiredResource.MemRequired),
			MemoryReservation: int64(requiredResource.MemRequired),
			NanoCPUs: int64(requiredResource.CpuRequired),
		},
		//StorageOpt: map[string]string{
		//	"size": fmt.Sprintf("%dG", requiredResource.DiskRequired / (1024 * 1024 * 1024)),
		//},
    }

	networkConfig := &network.NetworkingConfig{}

	containerName := fmt.Sprintf("ssh-enabled-container-%v",time.Now().Unix())
	resp, err := c.Client.ContainerCreate(context.Background(), containerConfig, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		log.Fatalf("Error creating container: %v", err)
	}

	fmt.Printf("Created container %s\n", resp.ID)

	err = c.Client.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		log.Fatalf("Error starting container: %v", err)
	}

	fmt.Printf("Container %s is running and SSH is available on port %v.\n", resp.ID , sshPort)
	return nil

}

func(c DockerClient) GetContainerList() ([]ContainerInfo , error){
	containerList , err := c.Client.ContainerList(context.Background() , container.ListOptions{})
	if err != nil {
		return nil , err
	}
	containerInfoList := make([]ContainerInfo , len(containerList)) 
	for idx , val := range(containerList) {
		containerInfoList[idx] = ContainerInfo{
			containerId: val.ID,
			image: val.Image,
			imageId: val.ImageID,
			ports: convertPort(val.Ports),
			state: val.State,
			status: val.Status,			
		}
	}
	return containerInfoList , err
}

func(c DockerClient) StopDockerContainer(containerId string) (error) {
	err := c.Client.ContainerStop(context.Background(),containerId,container.StopOptions{})
	if err != nil {
		return fmt.Errorf("error stoping container")
	}
	return err
}

func convertPort(ports []types.Port) []Port {
	portList := make([]Port,len(ports))
	for idx , port := range(ports) {
		portList[idx] = Port{
			portIP: port.IP,
			privatePort: port.PrivatePort,
			publicPort: port.PublicPort,
			portType: port.Type,
		}
	}

	return portList
}