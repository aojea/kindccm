package kindccm

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
)

type KindccmLoadBalancer struct {
	kubeClient *kubernetes.Clientset
}

var _ cloudprovider.LoadBalancer = &KindccmLoadBalancer{}

func NewKindccmLoadBalancer(kubeClient *kubernetes.Clientset) cloudprovider.LoadBalancer {
	return &KindccmLoadBalancer{kubeClient}
}

// GetLoadBalancer is a stub implementation of LoadBalancer.GetLoadBalancer.
func (k *KindccmLoadBalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	status := &v1.LoadBalancerStatus{}
	status.Ingress = []v1.LoadBalancerIngress{{IP: k.ExternalIP.String()}}

	return status, k.Exists, k.Err
}

// GetLoadBalancerName is a stub implementation of LoadBalancer.GetLoadBalancerName.
func (k *KindccmLoadBalancer) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
	// TODO: replace DefaultLoadBalancerName to generate more meaningful loadbalancer names.
	return cloudprovider.DefaultLoadBalancerName(service)
}

// EnsureLoadBalancer is a test-spy implementation of LoadBalancer.EnsureLoadBalancer.
// It adds an entry "create" into the internal method call record.
func (k *KindccmLoadBalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	// TODO: https://github.com/elsonrodriguez/minikube-lb-patch/blob/master/main.go
	if k.Balancers == nil {
		k.Balancers = make(map[string]Balancer)
	}

	name := k.GetLoadBalancerName(ctx, clusterName, service)
	spec := service.Spec

	zone, err := k.GetZone(context.TODO())
	if err != nil {
		return nil, err
	}
	region := zone.Region

	k.Balancers[name] = Balancer{name, region, spec.LoadBalancerIP, spec.Ports, nodes}

	status := &v1.LoadBalancerStatus{}
	status.Ingress = []v1.LoadBalancerIngress{{IP: k.ExternalIP.String()}}

	return status, k.Err
}

// UpdateLoadBalancer is a test-spy implementation of LoadBalancer.UpdateLoadBalancer.
// It adds an entry "update" into the internal method call record.
func (k *KindccmLoadBalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	k.addCall("update")
	k.UpdateCalls = append(k.UpdateCalls, UpdateBalancerCall{service, nodes})
	return k.Err
}

// EnsureLoadBalancerDeleted is a test-spy implementation of LoadBalancer.EnsureLoadBalancerDeleted.
// It adds an entry "delete" into the internal method call record.
func (k *KindccmLoadBalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	k.addCall("delete")
	return k.Err
}
