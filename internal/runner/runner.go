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
	ContainerId string
	State       string
	Status      string
	Image       string
	ImageId     string
	Ports       []Port
}

type Port struct {
	portIP      string
	privatePort uint16
	publicPort  uint16
	portType    string
}

type PortMapping struct {
    HostPort      string // Port on the host
    ContainerPort string // Port inside the container
}

func InitDockerClient() (*DockerClient, error) {
	apiClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		Client: apiClient,
	}, nil
}

func (c DockerClient) PullSSHEnabledUbunutImage() error {

	log.Println("pulling docker image")
	reader, err := c.Client.ImagePull(context.Background(), sshImage, image.PullOptions{})
	if err != nil {
		log.Println("Error pulling image:", err)
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
			log.Printf("\r%v\n", status)
		}
	}
	return nil
}

func (c DockerClient) StartSSHContainer(portMappings []PortMapping, requiredResource rtypes.Unit) (string, error) {
    exposedPorts := nat.PortSet{}
    portBindings := nat.PortMap{}

    for _, mapping := range portMappings {
        err := manager.CheckPortAvailable(mapping.HostPort)
        if err != nil {
            return "", fmt.Errorf("PORT %v is already taken", mapping.HostPort)
        }
        portKey := nat.Port(fmt.Sprintf("%s/tcp", mapping.ContainerPort))
        exposedPorts[portKey] = struct{}{}
        portBindings[portKey] = []nat.PortBinding{
            {
                HostIP:   "0.0.0.0",
                HostPort: mapping.HostPort,
            },
        }
    }

    containerConfig := &container.Config{
        Image:        sshImage,
        ExposedPorts: exposedPorts,
    }

    hostConfig := &container.HostConfig{
        PortBindings: portBindings,
        Resources: container.Resources{
            Memory:            int64(requiredResource.MemRequired),
            MemoryReservation: int64(requiredResource.MemRequired),
            NanoCPUs:          int64(requiredResource.CpuRequired),
        },
    }

    networkConfig := &network.NetworkingConfig{}
    containerName := fmt.Sprintf("ssh-enabled-container-%v", time.Now().Unix())
    resp, err := c.Client.ContainerCreate(context.Background(), containerConfig, hostConfig, networkConfig, nil, containerName)
    if err != nil {
        log.Printf("Error creating container: %v", err)
        return "", err
    }

    log.Printf("Created container %s\n", resp.ID)

    err = c.Client.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
    if err != nil {
        log.Printf("Error starting container: %v", err)
        return "", nil
    }

    log.Printf("Container %s is running and ports are mapped as specified.\n", resp.ID)
    return resp.ID, nil
}
func (c DockerClient) GetContainerList() ([]ContainerInfo, error) {
	containerList, err := c.Client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}
	containerInfoList := make([]ContainerInfo, len(containerList))
	for idx, val := range containerList {
		containerInfoList[idx] = ContainerInfo{
			ContainerId: val.ID,
			Image:       val.Image,
			ImageId:     val.ImageID,
			Ports:       convertPort(val.Ports),
			State:       val.State,
			Status:      val.Status,
		}
	}
	return containerInfoList, err
}

func (c DockerClient) StopDockerContainer(containerId string) error {
	fmt.Printf("%v", containerId)
	err := c.Client.ContainerStop(context.Background(), containerId, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("error stoping container : %v", err)
	}
	return err
}

func (c DockerClient) DeleteDockerContainer(containerId string) error {
	err := c.Client.ContainerRemove(context.Background(), containerId, container.RemoveOptions{RemoveVolumes: true,
		Force: true})
	if err != nil {
		return err
	}
	return nil
}

func (c DockerClient) GetContainerStatus(containerId string) (string, error) {
	resp, err := c.Client.ContainerStats(context.Background(), containerId, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}

func convertPort(ports []types.Port) []Port {
	portList := make([]Port, len(ports))
	for idx, port := range ports {
		portList[idx] = Port{
			portIP:      port.IP,
			privatePort: port.PrivatePort,
			publicPort:  port.PublicPort,
			portType:    port.Type,
		}
	}

	return portList
}

func (p Port) ToString() string {
	portStr := fmt.Sprintf("%v,%v,%v,%v", p.portIP, p.portType, p.privatePort, p.publicPort)
	return portStr
}
