package kindccm

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

const (
	ProviderName = "kindccm"
)

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, newKindccmCloudProvider)
}

type KindccmCloudProvider struct {
	lb cloudprovider.LoadBalancer
}

var _ cloudprovider.Interface = &KindccmCloudProvider{}

func newKindccmCloudProvider() (cloudprovider.Interface, error) {

	cfg, err := rest.InClusterConfig()

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client config: %s", err.Error())
	}

	cl, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %s", err.Error())
	}

	return &KindccmCloudProvider{NewKindccmLoadBalancer(cl)}, nil
}

// LoadBalancer returns a loadbalancer interface. Also returns true if the interface is supported, false otherwise.
func (k *KindccmCloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return k.lb, true
}

// Instances returns an instances interface. Also returns true if the interface is supported, false otherwise.
func (k *KindccmCloudProvider) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

// Zones returns a zones interface. Also returns true if the interface is supported, false otherwise.
func (k *KindccmCloudProvider) Zones() (cloudprovider.Zones, bool) {
	return zones{}, true
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (k *KindccmCloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is supported.
func (k *KindccmCloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (k *KindccmCloudProvider) ProviderName() string {
	return ProviderName
}

// HasClusterID returns true if a ClusterID is required and set
func (k *KindccmCloudProvider) HasClusterID() bool {
	return false
}
