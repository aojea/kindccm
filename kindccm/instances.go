package kindccm

import (
	"context"
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
)

type KindccmInstance struct {
	kubeClient *kubernetes.Clientset
}

var _ cloudprovider.Clusters = &KindccmInstance{}

func NewKindccmInstance(kubeClient *kubernetes.Clientset) cloudprovider.Clusters {
	return &KindccmInstance{kubeClient}
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (k *KindccmInstance) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return cloudprovider.NotImplemented
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (k *KindccmInstance) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// NodeAddresses is a test-spy implementation of Instances.NodeAddresses.
// It adds an entry "node-addresses" into the internal method call record.
func (k *KindccmInstance) NodeAddresses(ctx context.Context, instance types.NodeName) ([]v1.NodeAddress, error) {
	f.addCall("node-addresses")
	f.addressesMux.Lock()
	defer f.addressesMux.Unlock()
	return f.Addresses, f.Err
}

// SetNodeAddresses sets the addresses for a node
func (k *KindccmInstance) SetNodeAddresses(nodeAddresses []v1.NodeAddress) {
	f.addressesMux.Lock()
	defer f.addressesMux.Unlock()
	f.Addresses = nodeAddresses
}

// NodeAddressesByProviderID is a test-spy implementation of Instances.NodeAddressesByProviderID.
// It adds an entry "node-addresses-by-provider-id" into the internal method call record.
func (k *KindccmInstance) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	f.addCall("node-addresses-by-provider-id")
	f.addressesMux.Lock()
	defer f.addressesMux.Unlock()
	return f.Addresses, f.Err
}

// InstanceID returns the cloud provider ID of the node with the specified Name, unless an entry
// for the node exists in ExtIDError, in which case it returns the desired error (to facilitate
// testing of error handling).
func (k *KindccmInstance) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	f.addCall("instance-id")

	err, ok := f.ExtIDErr[nodeName]
	if ok {
		return "", err
	}

	return f.ExtID[nodeName], nil
}

// InstanceType returns the type of the specified instance.
func (k *KindccmInstance) InstanceType(ctx context.Context, instance types.NodeName) (string, error) {
	f.addCall("instance-type")
	return f.InstanceTypes[instance], nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (k *KindccmInstance) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	f.addCall("instance-type-by-provider-id")
	return f.InstanceTypes[types.NodeName(providerID)], nil
}

// InstanceExistsByProviderID returns true if the instance with the given provider id still exists and is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (k *KindccmInstance) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	f.addCall("instance-exists-by-provider-id")
	return f.ExistsByProviderID, f.ErrByProviderID
}

// InstanceShutdownByProviderID returns true if the instances is in safe state to detach volumes
func (k *KindccmInstance) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	f.addCall("instance-shutdown-by-provider-id")
	return f.NodeShutdown, f.ErrShutdownByProviderID
}

// List is a test-spy implementation of Instances.List.
// It adds an entry "list" into the internal method call record.
func (k *KindccmInstance) List(filter string) ([]types.NodeName, error) {
	f.addCall("list")
	result := []types.NodeName{}
	for _, machine := range f.Machines {
		if match, _ := regexp.MatchString(filter, string(machine)); match {
			result = append(result, machine)
		}
	}
	return result, f.Err
}
