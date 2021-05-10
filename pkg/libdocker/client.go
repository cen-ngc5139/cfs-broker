package libdocker

import (
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockerapi "github.com/docker/docker/client"
	"k8s.io/klog"
)

// Interface is an abstract interface for testability.  It abstracts the interface of docker client.
type Interface interface {
	ListContainers(options dockertypes.ContainerListOptions) ([]dockertypes.Container, error)
	InspectContainer(id string) (*dockertypes.ContainerJSON, error)
	GetContainerStats(id string) (*dockertypes.StatsJSON, error)
}

func getDockerClient() (*dockerapi.Client, error) {
	//获取cli客户端对象
	return dockerapi.NewClientWithOpts(dockerapi.FromEnv)
}

func ConnectToDockerOrDie(requestTimeout time.Duration) Interface {
	client, err := getDockerClient()
	if err != nil {
		klog.Fatalf("Couldn't connect to docker: %v", err)
	}
	klog.Infof("Start docker client with request timeout=%v", requestTimeout)
	return newKubeDockerClient(client, requestTimeout)
}
