package libdocker

import (
	"context"
	"encoding/json"
	"time"

	dockertypes "github.com/docker/docker/api/types"

	dockerapi "github.com/docker/docker/client"
)

const (
	defaultTimeout = 2*time.Minute - 1*time.Second
)

type KubeDockerClient struct {
	Client  *dockerapi.Client `json:"client"`
	Timeout time.Duration     `json:"timeout"`
}

func newKubeDockerClient(dockerClient *dockerapi.Client, requestTimeout time.Duration) *KubeDockerClient {

	if requestTimeout == 0 {
		requestTimeout = defaultTimeout
	}

	k := &KubeDockerClient{
		Client:  dockerClient,
		Timeout: requestTimeout,
	}

	// Notice that this assumes that docker is running before kubelet is started.
	ctx, cancel := k.getTimeoutContext()
	defer cancel()
	dockerClient.NegotiateAPIVersion(ctx)

	return k
}

func (k *KubeDockerClient) ListContainers(options dockertypes.ContainerListOptions) ([]dockertypes.Container, error) {
	ctx, cancel := k.getTimeoutContext()
	defer cancel()
	containers, err := k.Client.ContainerList(ctx, options)
	if ctxErr := contextError(ctx); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (k *KubeDockerClient) InspectContainer(id string) (*dockertypes.ContainerJSON, error) {
	ctx, cancel := k.getTimeoutContext()
	defer cancel()
	inspect, err := k.Client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	return &inspect, nil
}

// GetContainerStats is currently only used for Windows container stats
func (k *KubeDockerClient) GetContainerStats(id string) (*dockertypes.StatsJSON, error) {

	ctx, cancel := k.getCancelableContext()
	defer cancel()

	response, err := k.Client.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(response.Body)
	var stats dockertypes.StatsJSON
	err = dec.Decode(&stats)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return &stats, nil
}

// getCancelableContext returns a new cancelable context. For long running requests without timeout, we use cancelable
// context to avoid potential resource leak, although the current implementation shouldn't leak resource.
func (k *KubeDockerClient) getCancelableContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// getTimeoutContext returns a new context with default request timeout
func (k *KubeDockerClient) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), k.Timeout)
}
