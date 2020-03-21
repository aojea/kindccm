package kindccm

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	cloudprovider "k8s.io/cloud-provider"
)

const ProviderName = "KindCCM"

// Balancer is a storage of balancer information
type Balancer struct {
	Name           string
	LoadBalancerIP string
	Ports          []v1.ServicePort
	Hosts          []*v1.Node
}

var _ cloudprovider.LoadBalancer = (*Cloud)(nil)

// Cloud is a implementation of Interface, LoadBalancer, Instances, and Routes.
type Cloud struct {
	kubeClient clientset.Interface
	ExternalIP net.IP
	Balancers  map[string]Balancer
	Lock       sync.Mutex
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, newKindCloudProvider)
}

func newKindCloudProvider(io.Reader) (cloudprovider.Interface, error) {
	cfg, err := rest.InClusterConfig()

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client config: %s", err.Error())
	}

	cl, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %s", err.Error())
	}

	return &Cloud{
		kubeClient: cl,
	}, nil
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider
func (k *Cloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (k *Cloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (k *Cloud) ProviderName() string {
	return ProviderName
}

// HasClusterID returns true if the cluster has a clusterID
func (k *Cloud) HasClusterID() bool {
	return true
}

// LoadBalancer returns KIND implementation of LoadBalancer.
func (k *Cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return k, true
}

// Instances is not implemented.
func (k *Cloud) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

// Zones is not implemented.
func (k *Cloud) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

// Routes is not implemented.
func (k *Cloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// LoadBalancer

// GetLoadBalancer is a stub implementation of LoadBalancer.GetLoadBalancer.
func (k *Cloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	status := &v1.LoadBalancerStatus{}
	status.Ingress = []v1.LoadBalancerIngress{{IP: k.ExternalIP.String()}}

	return status, true, nil
}

// GetLoadBalancerName is a stub implementation of LoadBalancer.GetLoadBalancerName.
func (k *Cloud) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
	// TODO: replace DefaultLoadBalancerName to generate more meaningful loadbalancer names.
	return cloudprovider.DefaultLoadBalancerName(service)
}

// EnsureLoadBalancer is a test-spy implementation of LoadBalancer.EnsureLoadBalancer.
// It adds an entry "create" into the internal method call record.
func (k *Cloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	if k.Balancers == nil {
		k.Balancers = make(map[string]Balancer)
	}

	name := k.GetLoadBalancerName(ctx, clusterName, service)
	spec := service.Spec

	k.Balancers[name] = Balancer{name, spec.LoadBalancerIP, spec.Ports, nodes}

	status := &v1.LoadBalancerStatus{}
	status.Ingress = []v1.LoadBalancerIngress{{IP: k.ExternalIP.String()}}

	return status, nil
}

// UpdateLoadBalancer is a test-spy implementation of LoadBalancer.UpdateLoadBalancer.
// It adds an entry "update" into the internal method call record.
func (k *Cloud) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {

	return nil
}

// EnsureLoadBalancerDeleted is a test-spy implementation of LoadBalancer.EnsureLoadBalancerDeleted.
// It adds an entry "delete" into the internal method call record.
func (k *Cloud) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	return nil
}
