package models

// PodSandboxMetadata holds all necessary information for building the sandbox name.
// The container runtime is encouraged to expose the metadata associated with the
// PodSandbox in its user interface for better user experience. For example,
// the runtime can construct a unique PodSandboxName based on the metadata.
type PodSandboxMetadata struct {
	// Pod name of the sandbox. Same as the pod name in the Pod ObjectMeta.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Pod UID of the sandbox. Same as the pod UID in the Pod ObjectMeta.
	Uid string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	// Pod namespace of the sandbox. Same as the pod namespace in the Pod ObjectMeta.
	Namespace string `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// Attempt number of creating the sandbox. Default: 0.
	Attempt              uint32   `protobuf:"varint,4,opt,name=attempt,proto3" json:"attempt,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// ContainerMetadata holds all necessary information for building the container
// name. The container runtime is encouraged to expose the metadata in its user
// interface for better user experience. E.g., runtime can construct a unique
// container name based on the metadata. Note that (name, attempt) is unique
// within a sandbox for the entire lifetime of the sandbox.
type ContainerMetadata struct {
	// Name of the container. Same as the container name in the PodSpec.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Attempt number of creating the container. Default: 0.
	Attempt   uint32 `protobuf:"varint,2,opt,name=attempt,proto3" json:"attempt,omitempty"`
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
}
