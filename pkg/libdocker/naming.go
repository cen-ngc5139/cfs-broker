package libdocker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ghostbaby/cfs-broker/pkg/models"
)

const (
	// kubePrefix is used to identify the containers/sandboxes on the node managed by kubelet
	kubePrefix = "k8s"
	// sandboxContainerName is a string to include in the docker container so
	// that users can easily identify the sandboxes.
	sandboxContainerName = "POD"
	// Delimiter used to construct docker container names.
	nameDelimiter = "_"
	// DockerImageIDPrefix is the prefix of image id in container status.
	DockerImageIDPrefix = "docker://"
	// DockerPullableImageIDPrefix is the prefix of pullable image id in container status.
	DockerPullableImageIDPrefix = "docker-pullable://"
)

func parseUint32(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// TODO: Evaluate whether we should rely on labels completely.
func ParseSandboxName(name string) (*models.PodSandboxMetadata, error) {
	// Docker adds a "/" prefix to names. so trim it.
	name = strings.TrimPrefix(name, "/")

	parts := strings.Split(name, nameDelimiter)
	// Tolerate the random suffix.
	// TODO(random-liu): Remove 7 field case when docker 1.11 is deprecated.
	if len(parts) != 6 && len(parts) != 7 {
		return nil, fmt.Errorf("failed to parse the sandbox name: %q", name)
	}
	if parts[0] != kubePrefix {
		return nil, fmt.Errorf("container is not managed by kubernetes: %q", name)
	}

	attempt, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("failed to parse the sandbox name %q: %v", name, err)
	}

	return &models.PodSandboxMetadata{
		Name:      parts[2],
		Namespace: parts[3],
		Uid:       parts[4],
		Attempt:   attempt,
	}, nil
}

// TODO: Evaluate whether we should rely on labels completely.
func ParseContainerName(name string) (*models.ContainerMetadata, error) {
	// Docker adds a "/" prefix to names. so trim it.
	name = strings.TrimPrefix(name, "/")

	parts := strings.Split(name, nameDelimiter)
	// Tolerate the random suffix.
	// TODO(random-liu): Remove 7 field case when docker 1.11 is deprecated.
	if len(parts) != 6 && len(parts) != 7 {
		return nil, fmt.Errorf("failed to parse the container name: %q", name)
	}
	if parts[0] != kubePrefix {
		return nil, fmt.Errorf("container is not managed by kubernetes: %q", name)
	}

	attempt, err := parseUint32(parts[5])
	if err != nil {
		return nil, fmt.Errorf("failed to parse the container name %q: %v", name, err)
	}

	return &models.ContainerMetadata{
		Name:      parts[1],
		Attempt:   attempt,
		Namespace: parts[3],
		PodName:   parts[2],
	}, nil
}
